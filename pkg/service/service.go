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
	"github.com/SENERGY-Platform/api-docs-provider/pkg/models"
	srv_info_hdl "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl"
)

type Service struct {
	swaggerService
	asyncapiService
	srv_info_hdl.SrvInfoHandler
}

func New(swaggerSrv swaggerService, asyncapiSrv asyncapiService, srvInfoHdl srv_info_hdl.SrvInfoHandler) *Service {
	return &Service{
		swaggerService:  swaggerSrv,
		asyncapiService: asyncapiSrv,
		SrvInfoHandler:  srvInfoHdl,
	}
}

func (s *Service) HealthCheck(ctx context.Context) error {
	if _, err := s.SwaggerListStorage(ctx); err != nil {
		return models.NewInternalError(err)
	}
	if _, err := s.AsyncapiListStorage(ctx); err != nil {
		return models.NewInternalError(err)
	}
	return nil
}
