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
	"github.com/SENERGY-Platform/api-docs-provider/pkg/models"
	srv_util "github.com/SENERGY-Platform/api-docs-provider/pkg/service/util"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util/slog_attr"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"path"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"
)

func (s *Service) SwaggerPeriodicProcurement(ctx context.Context, interval, delay time.Duration) error {
	logger.Info("starting periodic procurement")
	var lErr error
	defer func() {
		if r := recover(); r != nil {
			lErr = fmt.Errorf("%s", r)
			logger.Error("periodic procurement panicked", slog_attr.StackTraceKey, string(debug.Stack()))
		}
		logger.Info("periodic procurement halted")
	}()
	timer := time.NewTimer(delay)
	loop := true
	for loop {
		select {
		case <-timer.C:
			err := s.SwaggerRefreshDocs(ctx)
			if err != nil {
				var rbe *lib_models.ResourceBusyError
				if !errors.As(err, &rbe) {
					logger.Error("procurement failed", attributes.ErrorKey, err)
				}
			}
			timer.Reset(interval)
		case <-ctx.Done():
			loop = false
			logger.Info("stopping periodic procurement")
			break
		}
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	return lErr
}

func (s *Service) SwaggerRefreshDocs(ctx context.Context) error {
	if !s.mu.TryLock() {
		return lib_models.NewResourceBusyError(errors.New("procurement running"))
	}
	defer s.mu.Unlock()
	services, err := s.discoveryHdl.GetServices(ctx)
	if err != nil {
		return lib_models.NewInternalError(err)
	}
	wg := &sync.WaitGroup{}
	for _, service := range services {
		if err = ctx.Err(); err != nil {
			break
		}
		if len(service.ExtPaths) > 0 {
			wg.Add(1)
			go s.handleService(ctx, wg, service)
		}
	}
	wg.Wait()
	if err = s.cleanOldServices(ctx, services); err != nil {
		logger.Error("removing old docs failed", attributes.ErrorKey, err, slog_attr.RequestIDKey, util.GetReqID(ctx))
	}
	return nil
}

func (s *Service) cleanOldServices(ctx context.Context, services map[string]models.Service) error {
	storedServices, err := s.storageHdl.List(ctx)
	if err != nil {
		return err
	}
	servicesSet := make(map[string]struct{})
	for _, service := range services {
		for _, extPath := range service.ExtPaths {
			servicesSet[getStorageID(service.ID, extPath)] = struct{}{}
		}
	}
	for _, service := range storedServices {
		if _, ok := servicesSet[service.ID]; !ok {
			if err = s.storageHdl.Delete(ctx, service.ID); err != nil {
				logger.Error("removing old doc failed", attributes.ErrorKey, err, slog_attr.RequestIDKey, util.GetReqID(ctx))
			}
		}
	}
	return nil
}

func (s *Service) handleService(ctx context.Context, wg *sync.WaitGroup, service models.Service) {
	defer wg.Done()
	ctxWt, cf := context.WithTimeout(ctx, s.timeout)
	defer cf()
	reqID := util.GetReqID(ctx)
	logger.Debug("probing host", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, slog_attr.RequestIDKey, reqID)
	doc, err := s.docClt.GetDoc(ctxWt, service.Protocol, service.Host, service.Port)
	if err != nil {
		logger.Debug("probing host failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
	var tmp map[string]json.RawMessage
	if err := json.Unmarshal(doc, &tmp); err != nil {
		logger.Error("unmarshalling doc failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
	if err = validateSwaggerKeys(tmp); err != nil {
		logger.Error("validating keys failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
	sInfo, err := getSwaggerInfo(tmp)
	if err != nil {
		logger.Error("extracting info failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
	sPaths, err := getSwaggerPaths(tmp)
	if err != nil {
		logger.Error("extracting paths failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
	if err = s.setSwaggerHostAndSchemes(tmp); err != nil {
		logger.Error("setting swagger host and schemes failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
	for _, extPath := range service.ExtPaths {
		if err = s.setSwaggerBasePath(tmp, extPath); err != nil {
			logger.Error("setting swagger base path failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, slog_attr.BasePathKey, extPath, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
			continue
		}
		b, err := json.Marshal(tmp)
		if err != nil {
			logger.Error("marshaling doc failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, slog_attr.BasePathKey, extPath, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
			continue
		}
		args := [][2]string{
			{titleArgKey, sInfo.Title},
			{versionArgKey, sInfo.Version},
			{descriptionArgKey, sInfo.Description},
			{basePathArgKey, extPath},
		}
		for _, route := range newRoutes(sPaths, extPath) {
			args = append(args, [2]string{routeArgKey, route})
		}
		if err = s.storageHdl.Write(ctx, getStorageID(service.ID, extPath), args, b); err != nil {
			logger.Error("writing doc failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, slog_attr.BasePathKey, extPath, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
			continue
		}
	}
}

func validateSwaggerKeys(tmp map[string]json.RawMessage) error {
	if !srv_util.CheckForKeys(tmp, swaggerV2Keys) && !srv_util.CheckForKeys(tmp, swaggerV3Keys) {
		return errors.New("missing required keys")
	}
	return nil
}

func getSwaggerInfo(tmp map[string]json.RawMessage) (swaggerInfo, error) {
	raw, ok := tmp[swaggerInfoKey]
	if !ok {
		return swaggerInfo{}, errors.New("missing key")
	}
	var info swaggerInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return swaggerInfo{}, err
	}
	return info, nil
}

func (s *Service) setSwaggerHostAndSchemes(tmp map[string]json.RawMessage) error {
	b, err := json.Marshal(s.apiGtwHost)
	if err != nil {
		return err
	}
	tmp[swaggerHostKey] = b
	if _, ok := tmp[swaggerSchemesKey]; !ok {
		b, err = json.Marshal([]string{"https"})
		if err != nil {
			return err
		}
		tmp[swaggerSchemesKey] = b
	}
	return nil
}

func (s *Service) setSwaggerBasePath(tmp map[string]json.RawMessage, basePath string) error {
	b, err := json.Marshal(basePath)
	if err != nil {
		return err
	}
	tmp[swaggerBasePathKey] = b
	return nil
}

func newRoutes(pathsMap map[string]map[string]json.RawMessage, basePath string) []string {
	var routes []string
	for pth, obj := range pathsMap {
		for mth := range obj {
			routes = append(routes, fmt.Sprintf("%s%s%s", path.Join(basePath, pth), routeDelimiter, mth))
		}
	}
	slices.Sort(routes)
	return routes
}

func getStorageID(srvID, extPath string) string {
	return srvID + strings.Replace(extPath, "/", "_", -1)
}
