package config

import (
	envldr "github.com/SENERGY-Platform/go-env-loader"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	"github.com/y-du/go-log-level/level"
	"reflect"
	"strings"
)

var typeParser = map[reflect.Type]envldr.Parser{
	reflect.TypeOf(level.Off):  sb_logger.LevelParser,
	reflect.TypeOf([]string{}): listParser,
}

func listParser(_ reflect.Type, val string, _ []string, kwParams map[string]string) (interface{}, error) {
	sep, ok := kwParams["sep"]
	if !ok {
		sep = ","
	}
	return strings.Split(val, sep), nil
}
