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
	srv_info_hdl "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl"
	srv_info_lib "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl/lib"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
)

type Service struct {
	swaggerHdl SwaggerHandler
	srvInfoHdl srv_info_hdl.SrvInfoHandler
}

func New(swaggerHdl SwaggerHandler, srvInfoHdl srv_info_hdl.SrvInfoHandler) *Service {
	return &Service{
		swaggerHdl: swaggerHdl,
		srvInfoHdl: srvInfoHdl,
	}
}

func (s *Service) SwaggerDocs(ctx context.Context, userToken string, userRoles []string) ([]map[string]json.RawMessage, error) {
	return s.swaggerHdl.GetDocs(ctx, userToken, userRoles)
}

func (s *Service) SwaggerStorageRefresh(ctx context.Context) error {
	return s.swaggerHdl.RefreshStorage(ctx)
}

func (s *Service) SwaggerStorageList(ctx context.Context) ([]models.StorageData, error) {
	return s.swaggerHdl.ListStorage(ctx)
}

func (s *Service) SrvInfo(_ context.Context) srv_info_lib.SrvInfo {
	return s.srvInfoHdl.GetInfo()
}

func (s *Service) HealthCheck(ctx context.Context) error {
	if _, err := s.swaggerHdl.ListStorage(ctx); err != nil {
		return models.NewInternalError(err)
	}
	return nil
}
