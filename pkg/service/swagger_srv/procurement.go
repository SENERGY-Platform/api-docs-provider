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
	"runtime/debug"
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
	for _, service := range storedServices {
		if _, ok := services[service.ID]; !ok {
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
	if err = validateDoc(doc); err != nil {
		logger.Warn("validating doc failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
	title, version, err := getSwaggerInfo(doc)
	if err != nil {
		logger.Warn("extracting info failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
	args := [][2]string{
		{titleArgKey, title},
		{versionArgKey, version},
	}
	for _, path := range service.ExtPaths {
		args = append(args, [2]string{extPathArgKey, path})
	}
	if err = s.storageHdl.Write(ctx, service.ID, args, doc); err != nil {
		logger.Error("writing doc failed", slog_attr.HostKey, service.Host, slog_attr.PortKey, service.Port, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return
	}
}

func validateDoc(doc []byte) error {
	var tmp map[string]json.RawMessage
	if err := json.Unmarshal(doc, &tmp); err != nil {
		return err
	}
	if !srv_util.CheckForKeys(tmp, swaggerV2Keys) && !srv_util.CheckForKeys(tmp, swaggerV3Keys) {
		return errors.New("missing required keys")
	}
	return nil
}

func getSwaggerInfo(doc []byte) (string, string, error) {
	var info swaggerInfo
	if err := json.Unmarshal(doc, &info); err != nil {
		return "", "", err
	}
	return info.Info.Title, info.Info.Version, nil
}
