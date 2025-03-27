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

package service

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/go-service-base/structured-logger/attributes"
	srv_info_hdl "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl"
	srv_info_lib "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl/lib"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/doc_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/ladon_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util/slog_attr"
	"slices"
	"strings"
	"sync"
	"time"
)

type Service struct {
	storageHdl    StorageHandler
	discoveryHdl  DiscoveryHandler
	srvInfoHdl    srv_info_hdl.SrvInfoHandler
	docClt        doc_clt.ClientItf
	ladonClt      ladon_clt.ClientItf
	timeout       time.Duration
	apiGtwHost    string
	adminRoleName string
	mu            sync.Mutex
}

func New(storageHdl StorageHandler, discoveryHdl DiscoveryHandler, srvInfoHdl srv_info_hdl.SrvInfoHandler, docClt doc_clt.ClientItf, ladonClt ladon_clt.ClientItf, timeout time.Duration, apiGtwHost string, adminRoleName string) *Service {
	return &Service{
		storageHdl:    storageHdl,
		discoveryHdl:  discoveryHdl,
		srvInfoHdl:    srvInfoHdl,
		docClt:        docClt,
		ladonClt:      ladonClt,
		timeout:       timeout,
		apiGtwHost:    apiGtwHost,
		adminRoleName: adminRoleName,
	}
}

func (s *Service) SwaggerDocs(ctx context.Context, userToken string, userRoles []string) ([]map[string]json.RawMessage, error) {
	if userToken == "" && len(userRoles) == 0 {
		return []map[string]json.RawMessage{}, nil
	}
	data, err := s.storageHdl.List(ctx)
	if err != nil {
		return []map[string]json.RawMessage{}, models.NewInternalError(err)
	}
	reqID := util.GetReqID(ctx)
	isAdmin := stringInSlice(s.adminRoleName, userRoles)
	var docWrappers []docWrapper
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for _, item := range data {
		wg.Add(1)
		go func(id string, extPaths []string) {
			defer wg.Done()
			logger.Debug("reading swagger doc", slog_attr.ExternalPathsKey, extPaths, slog_attr.RequestIDKey, reqID)
			rawDoc, err := s.storageHdl.Read(ctx, id)
			if err != nil {
				logger.Error("reading swagger doc failed", slog_attr.ExternalPathsKey, extPaths, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
				return
			}
			for _, basePath := range extPaths {
				logger.Debug("transforming swagger doc'", slog_attr.BasePathKey, basePath, slog_attr.RequestIDKey, reqID)
				doc, err := s.transformDoc(rawDoc, basePath)
				if err != nil {
					logger.Error("transforming swagger doc failed", slog_attr.BasePathKey, basePath, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
					continue
				}
				if !isAdmin {
					logger.Debug("filtering swagger doc", slog_attr.BasePathKey, basePath, slog_attr.RequestIDKey, reqID)
					ok, err := s.filterDoc(ctx, doc, userToken, userRoles, basePath)
					if err != nil {
						logger.Error("filtering swagger doc failed", slog_attr.BasePathKey, basePath, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
						continue
					}
					if !ok {
						continue
					}
				}
				mu.Lock()
				docWrappers = append(docWrappers, docWrapper{basePath: basePath, doc: doc})
				mu.Unlock()
				logger.Debug("appended swagger doc", slog_attr.BasePathKey, basePath, slog_attr.RequestIDKey, reqID)
			}
		}(item.ID, item.ExtPaths)
	}
	wg.Wait()
	slices.SortStableFunc(docWrappers, func(a, b docWrapper) int {
		return strings.Compare(a.basePath, b.basePath)
	})
	docs := make([]map[string]json.RawMessage, 0, len(docWrappers))
	for _, dw := range docWrappers {
		docs = append(docs, dw.doc)
	}
	return docs, nil
}

func (s *Service) ListStorage(ctx context.Context) ([]models.StorageData, error) {
	items, err := s.storageHdl.List(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Service) HealthCheck(ctx context.Context) error {
	if _, err := s.storageHdl.List(ctx); err != nil {
		return models.NewInternalError(err)
	}
	return nil
}

func (s *Service) SrvInfo(_ context.Context) srv_info_lib.SrvInfo {
	return s.srvInfoHdl.GetInfo()
}

func (s *Service) transformDoc(rawDoc []byte, basePath string) (map[string]json.RawMessage, error) {
	var doc map[string]json.RawMessage
	err := json.Unmarshal(rawDoc, &doc)
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	b, err := json.Marshal(s.apiGtwHost)
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	doc[swaggerHostKey] = b
	b, err = json.Marshal(basePath)
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	doc[swaggerBasePathKey] = b
	if _, ok := doc[swaggerSchemesKey]; !ok {
		b, err = json.Marshal([]string{"https"})
		if err != nil {
			return nil, models.NewInternalError(err)
		}
		doc[swaggerSchemesKey] = b
	}
	return doc, nil
}

func stringInSlice(a string, sl []string) bool {
	for _, b := range sl {
		if b == a {
			return true
		}
	}
	return false
}
