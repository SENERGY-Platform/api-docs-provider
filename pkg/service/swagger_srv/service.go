/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package swagger_srv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	lib_models "github.com/SENERGY-Platform/api-docs-provider/lib/models"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/components/doc_clt"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/components/ladon_clt"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util/slog_attr"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"strings"
	"sync"
	"time"
)

const (
	basePathArgKey    = "base-path"
	routeArgKey       = "route"
	titleArgKey       = "title"
	versionArgKey     = "version"
	descriptionArgKey = "description"
)

const routeDelimiter = "|"

type Service struct {
	storageHdl    StorageHandler
	discoveryHdl  DiscoveryHandler
	docClt        doc_clt.ClientItf
	ladonClt      ladon_clt.ClientItf
	timeout       time.Duration
	apiGtwHost    string
	adminRoleName string
	mu            sync.Mutex
}

func New(storageHdl StorageHandler, discoveryHdl DiscoveryHandler, docClt doc_clt.ClientItf, ladonClt ladon_clt.ClientItf, timeout time.Duration, apiGtwHost string, adminRoleName string) *Service {
	return &Service{
		storageHdl:    storageHdl,
		discoveryHdl:  discoveryHdl,
		docClt:        docClt,
		ladonClt:      ladonClt,
		timeout:       timeout,
		apiGtwHost:    apiGtwHost,
		adminRoleName: adminRoleName,
	}
}

func (s *Service) SwaggerGetDocs(ctx context.Context, userToken string, userRoles []string) ([]map[string]json.RawMessage, error) {
	if userToken == "" && len(userRoles) == 0 {
		return []map[string]json.RawMessage{}, nil
	}
	storageItems, err := s.storageHdl.List(ctx)
	if err != nil {
		return []map[string]json.RawMessage{}, err
	}
	reqID := util.GetReqID(ctx)
	isAdmin := stringInSlice(s.adminRoleName, userRoles)
	var docs []map[string]json.RawMessage
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for _, item := range storageItems {
		if ctx.Err() != nil {
			return nil, lib_models.NewInternalError(ctx.Err())
		}
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			logger.Debug("reading doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
			rawDoc, err := s.storageHdl.Read(ctx, id)
			if err != nil {
				logger.Error("reading doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
				return
			}
			logger.Debug("unmarshalling doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
			var doc map[string]json.RawMessage
			if err = json.Unmarshal(rawDoc, &doc); err != nil {
				logger.Error("unmarshalling doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
				return
			}
			if !isAdmin {
				logger.Debug("filtering doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
				ok, err := s.filterDoc(ctx, doc, userToken, userRoles)
				if err != nil {
					logger.Error("filtering doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
					return
				}
				if !ok {
					return
				}
			}
			mu.Lock()
			docs = append(docs, doc)
			mu.Unlock()
			logger.Debug("appended doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
		}(item.ID)
	}
	wg.Wait()
	return docs, nil
}

func (s *Service) SwaggerGetDoc(ctx context.Context, id string, userToken string, userRoles []string) ([]byte, error) {
	reqID := util.GetReqID(ctx)
	logger.Debug("reading doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
	rawDoc, err := s.storageHdl.Read(ctx, id)
	if err != nil {
		logger.Error("reading doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
		return nil, err
	}
	logger.Debug("unmarshalling doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
	var tmp map[string]json.RawMessage
	if err = json.Unmarshal(rawDoc, &tmp); err != nil {
		logger.Error("unmarshalling doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
		return nil, lib_models.NewInternalError(err)
	}
	if !stringInSlice(s.adminRoleName, userRoles) {
		logger.Debug("filtering doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
		ok, err := s.filterDoc(ctx, tmp, userToken, userRoles)
		if err != nil {
			logger.Error("filtering doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
			return nil, lib_models.NewInternalError(err)
		}
		if !ok {
			return nil, lib_models.NewForbiddenErr(errors.New("no access rights"))
		}
	}
	doc, err := json.Marshal(tmp)
	if err != nil {
		return nil, lib_models.NewInternalError(err)
	}
	return doc, nil
}

func (s *Service) SwaggerListStorage(ctx context.Context, userToken string, userRoles []string) ([]lib_models.SwaggerItem, error) {
	if userToken == "" && len(userRoles) == 0 {
		return nil, nil
	}
	storageItems, err := s.storageHdl.List(ctx)
	if err != nil {
		return nil, err
	}
	reqID := util.GetReqID(ctx)
	var swaggerItems []lib_models.SwaggerItem
	if stringInSlice(s.adminRoleName, userRoles) {
		for _, storageItem := range storageItems {
			swaggerItems = append(swaggerItems, newSwaggerItem(storageItem))
		}
	} else {
		wg := sync.WaitGroup{}
		mu := sync.Mutex{}
		for _, storageItem := range storageItems {
			if ctx.Err() != nil {
				return nil, lib_models.NewInternalError(ctx.Err())
			}
			wg.Add(1)
			go func(sData models.StorageData) {
				defer wg.Done()
				swaggerItem := newSwaggerItem(sData)
				logger.Debug("getting routes", slog_attr.IDKey, swaggerItem.ID, slog_attr.BasePathKey, swaggerItem.BasePath, slog_attr.RequestIDKey, reqID)
				routes, err := getRoutes(sData.Args)
				if err != nil {
					logger.Error("getting routes failed", slog_attr.IDKey, swaggerItem.ID, slog_attr.BasePathKey, swaggerItem.BasePath, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
					return
				}
				logger.Debug("checking routes", slog_attr.IDKey, swaggerItem.ID, slog_attr.BasePathKey, swaggerItem.BasePath, slog_attr.RequestIDKey, reqID)
				ok, err := s.checkRoutes(ctx, userToken, userRoles, routes)
				if err != nil {
					logger.Error("checking routes failed", slog_attr.IDKey, swaggerItem.ID, slog_attr.BasePathKey, swaggerItem.BasePath, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
					return
				}
				if ok {
					mu.Lock()
					swaggerItems = append(swaggerItems, swaggerItem)
					mu.Unlock()
				}
			}(storageItem)
		}
	}
	return swaggerItems, nil
}

func (s *Service) checkRoutes(ctx context.Context, userToken string, userRoles []string, routes map[string][]string) (bool, error) {
	if len(routes) == 0 {
		return true, nil
	}
	if userToken != "" {
		pathMethodMap := make(map[string][]string)
		for pth, methods := range routes {
			pathMethodMap[pth] = methods
		}
		ctxWt, cf := context.WithTimeout(ctx, s.timeout)
		defer cf()
		accessPolicies, err := s.ladonClt.GetUserAccessPolicy(ctxWt, userToken, pathMethodMap)
		if err != nil {
			return false, err
		}
		return len(accessPolicies) > 0, nil
	} else {
		for pth, methods := range routes {
			for _, method := range methods {
				for _, role := range userRoles {
					ok, err := s.getAccessPolicyByRole(ctx, pth, role, method)
					if err != nil {
						return false, err
					}
					if ok {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

func stringInSlice(a string, sl []string) bool {
	for _, b := range sl {
		if b == a {
			return true
		}
	}
	return false
}

func getRoutes(args [][2]string) (map[string][]string, error) {
	routes := make(map[string][]string)
	for _, arg := range args {
		if arg[0] == routeArgKey {
			parts := strings.Split(arg[1], routeDelimiter)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid: %s", arg[1])
			}
			route := routes[parts[0]]
			route = append(route, parts[1])
			routes[parts[0]] = route
		}
	}
	return routes, nil
}

func newSwaggerItem(sd models.StorageData) lib_models.SwaggerItem {
	si := lib_models.SwaggerItem{
		ID: sd.ID,
	}
	for _, arg := range sd.Args {
		switch arg[0] {
		case titleArgKey:
			si.Title = arg[1]
		case versionArgKey:
			si.Version = arg[1]
		case descriptionArgKey:
			si.Description = arg[1]
		case basePathArgKey:
			si.BasePath = arg[1]
		}
	}
	return si
}
