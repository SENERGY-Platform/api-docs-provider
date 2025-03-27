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

package config

import (
	sb_config_hdl "github.com/SENERGY-Platform/go-service-base/config-hdl"
	sb_config_types "github.com/SENERGY-Platform/go-service-base/config-hdl/types"
	"github.com/SENERGY-Platform/go-service-base/struct-logger"
	"time"
)

type KongConfig struct {
	User     string                 `json:"user" env_var:"KONG_USER"`
	Password sb_config_types.Secret `json:"password" env_var:"KONG_PASSWORD"`
	BaseURL  string                 `json:"base_url" env_var:"KONG_BASE_URL"`
}

type ProcurementConfig struct {
	SwaggerDocPath string        `json:"swagger_doc_path" env_var:"SWAGGER_DOC_PATH"`
	Interval       time.Duration `json:"interval" env_var:"PROCUREMENT_INTERVAL"`
}

type FilterConfig struct {
	LadonBaseUrl  string `json:"ladon_base_url" env_var:"LADON_BASE_URL"`
	AdminRoleName string `json:"admin_role_name" env_var:"ADMIN_ROLE_NAME"`
}

type DiscoveryConfig struct {
	Kong          KongConfig `json:"kong" env_var:"KONG_CONFIG"`
	HostBlacklist []string   `json:"host_blacklist" env_var:"DISCOVERY_HOST_BLACKLIST" env_params:"sep=,"`
}

type Config struct {
	ServerPort    int                  `json:"server_port" env_var:"SERVER_PORT"`
	Logger        struct_logger.Config `json:"logger" env_var:"LOGGER_CONFIG"`
	WorkdirPath   string               `json:"workdir_path" env_var:"WORKDIR_PATH"`
	ApiGateway    string               `json:"api_gateway" env_var:"API_GATEWAY"`
	Discovery     DiscoveryConfig      `json:"discovery" env_var:"DISCOVERY_CONFIG"`
	Procurement   ProcurementConfig    `json:"procurement" env_var:"PROCUREMENT_CONFIG"`
	Filter        FilterConfig         `json:"filter" env_var:"FILTER_CONFIG"`
	HttpTimeout   time.Duration        `json:"http_timeout" env_var:"HTTP_TIMEOUT"`
	HttpAccessLog bool                 `json:"http_access_log" env_var:"HTTP_ACCESS_LOG"`
}

func New(path string) (*Config, error) {
	cfg := Config{
		ServerPort: 80,
		Logger: struct_logger.Config{
			Handler:    struct_logger.JsonHandlerSelector,
			Level:      struct_logger.LevelInfo,
			TimeFormat: time.RFC3339Nano,
			TimeUtc:    true,
		},
		WorkdirPath: "data",
		Procurement: ProcurementConfig{
			Interval: time.Hour * 6,
		},
		HttpTimeout: time.Second * 30,
	}
	err := sb_config_hdl.Load(&cfg, nil, envTypeParser, nil, path)
	return &cfg, err
}
