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
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
)

type DiscoveryHandler interface {
	GetServices(ctx context.Context) (map[string]models.Service, error)
}

type StorageHandler interface {
	List(ctx context.Context) ([]models.StorageData, error)
	Write(ctx context.Context, id string, args [][2]string, data []byte) error
	Read(ctx context.Context, id string) ([]byte, error)
	Delete(ctx context.Context, id string) error
}
