package config

import (
	config_hdl "github.com/SENERGY-Platform/go-service-base/config-hdl"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	"github.com/y-du/go-log-level/level"
	"reflect"
	"strings"
)

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
