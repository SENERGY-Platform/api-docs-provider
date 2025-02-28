package config

import (
	config_hdl "github.com/SENERGY-Platform/go-service-base/config-hdl"
	config_env_parser "github.com/SENERGY-Platform/go-service-base/config-hdl/env_parser"
	config_types "github.com/SENERGY-Platform/go-service-base/config-hdl/types"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	"github.com/y-du/go-log-level/level"
	"reflect"
	"strings"
)

var envTypeParser = []config_hdl.EnvTypeParser{
	config_types.SecretEnvTypeParser,
	config_env_parser.DurationEnvTypeParser,
	logLevelEnvTypeParser,
	listEnvTypeParser,
}

func logLevelEnvTypeParser() (reflect.Type, config_hdl.EnvParser) {
	return reflect.TypeOf(level.Off), sb_logger.LevelParser
}

func listEnvTypeParser() (reflect.Type, config_hdl.EnvParser) {
	return reflect.TypeOf([]string{}), listEnvParser
}

func listEnvParser(_ reflect.Type, val string, _ []string, kwParams map[string]string) (interface{}, error) {
	sep, ok := kwParams["sep"]
	if !ok {
		sep = ","
	}
	return strings.Split(val, sep), nil
}
