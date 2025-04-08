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
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util/slog_attr"
	"runtime/debug"
	"sync"
	"time"
)

func (s *Service) RunPeriodicProcurement(ctx context.Context, interval time.Duration) error {
	logger.Info("starting periodic procurement")
	var lErr error
	defer func() {
		if r := recover(); r != nil {
			lErr = fmt.Errorf("%s", r)
			logger.Error("periodic procurement panicked", slog_attr.StackTraceKey, string(debug.Stack()))
		}
		logger.Info("periodic procurement halted")
	}()
	timer := time.NewTimer(time.Microsecond)
	loop := true
	for loop {
		select {
		case <-timer.C:
			err := s.RefreshStorage(ctx)
			if err != nil {
				var rbe *models.ResourceBusyError
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

func (s *Service) RefreshStorage(ctx context.Context) error {
	if !s.mu.TryLock() {
		return models.NewResourceBusyError(errors.New("procurement running"))
	}
	defer s.mu.Unlock()
	services, err := s.discoveryHdl.GetServices(ctx)
	if err != nil {
		return models.NewInternalError(err)
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
	var args [][2]string
	for _, path := range service.ExtPaths {
		args = append(args, [2]string{extPathKey, path})
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
	if !checkForKeys(tmp, swaggerV2Keys) && !checkForKeys(tmp, swaggerV3Keys) {
		return errors.New("missing required keys")
	}
	return nil
}

func checkForKeys(doc map[string]json.RawMessage, keys []string) bool {
	c := 0
	for _, key := range keys {
		if _, ok := doc[key]; ok {
			c++
		}
	}
	return c == len(keys)
}
