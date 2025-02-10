package config

import (
	config_hdl "github.com/SENERGY-Platform/go-service-base/config-hdl"
	config_types "github.com/SENERGY-Platform/go-service-base/config-hdl/types"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	envldr "github.com/y-du/go-env-loader"
	"github.com/y-du/go-log-level/level"
	"reflect"
	"time"
)

type LoggerConfig struct {
	Level        level.Level `json:"level" env_var:"LOGGER_LEVEL"`
	Utc          bool        `json:"utc" env_var:"LOGGER_UTC"`
	Path         string      `json:"path" env_var:"LOGGER_PATH"`
	FileName     string      `json:"file_name" env_var:"LOGGER_FILE_NAME"`
	Terminal     bool        `json:"terminal" env_var:"LOGGER_TERMINAL"`
	Microseconds bool        `json:"microseconds" env_var:"LOGGER_MICROSECONDS"`
	Prefix       string      `json:"prefix" env_var:"LOGGER_PREFIX"`
}

type KongConfig struct {
	User     string              `json:"user" env_var:"KONG_USER"`
	Password config_types.Secret `json:"password" env_var:"KONG_PASSWORD"`
	BaseURL  string              `json:"base_url" env_var:"KONG_BASE_URL"`
}

type ProcurementConfig struct {
	SwaggerDocPath string `json:"swagger_doc_path" env_var:"SWAGGER_DOC_PATH"`
	Interval       int64  `json:"interval" env_var:"PROCUREMENT_INTERVAL"`
}

type FilterConfig struct {
	LadonBaseUrl  string `json:"ladon_base_url" env_var:"LADON_BASE_URL"`
	AdminRoleName string `json:"admin_role_name" env_var:"ADMIN_ROLE_NAME"`
}

type DiscoveryConfig struct {
	Kong          KongConfig `json:"kong" env_var:"KONG_CONFIG"`
	HostBlacklist []string   `json:"host_blacklist" env_var:"DISCOVERY_HOST_BLACKLIST"`
}

type Config struct {
	ServerPort  int               `json:"server_port" env_var:"SERVER_PORT"`
	Logger      LoggerConfig      `json:"logger" env_var:"LOGGER_CONFIG"`
	WorkdirPath string            `json:"workdir_path" env_var:"WORKDIR_PATH"`
	ApiGateway  string            `json:"api_gateway" env_var:"API_GATEWAY"`
	Discovery   DiscoveryConfig   `json:"discovery" env_var:"DISCOVERY_CONFIG"`
	Procurement ProcurementConfig `json:"procurement" env_var:"PROCUREMENT_CONFIG"`
	Filter      FilterConfig      `json:"filter" env_var:"FILTER_CONFIG"`
	HttpTimeout int64             `json:"http_timeout" env_var:"HTTP_TIMEOUT"`
}

func New(path string) (*Config, error) {
	cfg := Config{
		ServerPort: 80,
		Logger: LoggerConfig{
			Level:        level.Warning,
			Utc:          true,
			Microseconds: true,
			Terminal:     true,
		},
		WorkdirPath: "data",
		Procurement: ProcurementConfig{
			Interval: int64(time.Hour * 6),
		},
		HttpTimeout: int64(time.Second * 30),
	}
	err := config_hdl.Load(&cfg, nil, map[reflect.Type]envldr.Parser{reflect.TypeOf(level.Off): sb_logger.LevelParser}, nil, path)
	return &cfg, err
}
