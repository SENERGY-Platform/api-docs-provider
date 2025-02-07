package config

import (
	config_hdl "github.com/SENERGY-Platform/go-service-base/config-hdl"
	config_types "github.com/SENERGY-Platform/go-service-base/config-hdl/types"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	envldr "github.com/y-du/go-env-loader"
	"github.com/y-du/go-log-level/level"
	"reflect"
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

type ApiGatewayConfig struct {
	Host string `json:"host" env_var:"API_GATEWAY_HOST"`
	Port int    `json:"port" env_var:"API_GATEWAY_PORT"`
}

type Config struct {
	ServerPort   uint             `json:"server_port" env_var:"SERVER_PORT"`
	Logger       LoggerConfig     `json:"logger" env_var:"LOGGER_CONFIG"`
	Kong         KongConfig       `json:"kong" env_var:"KONG_CONFIG"`
	LadonBaseUrl string           `json:"ladon_base_url" env_var:"LADON_BASE_URL"`
	ApiGateway   ApiGatewayConfig `json:"api_gateway" env_var:"API_GATEWAY_CONFIG"`
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
	}
	err := config_hdl.Load(&cfg, nil, map[reflect.Type]envldr.Parser{reflect.TypeOf(level.Off): sb_logger.LevelParser}, nil, path)
	return &cfg, err
}
