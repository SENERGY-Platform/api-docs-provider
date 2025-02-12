package config

import (
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	envldr "github.com/y-du/go-env-loader"
	"github.com/y-du/go-log-level/level"
	"reflect"
	"strings"
	"time"
)

var typeParser = map[reflect.Type]envldr.Parser{
	reflect.TypeOf(level.Off):       sb_logger.LevelParser,
	reflect.TypeOf(time.Nanosecond): durationParser,
	reflect.TypeOf([]string{}):      listParser,
}

func durationParser(_ reflect.Type, val string, _ []string, _ map[string]string) (interface{}, error) {
	return time.ParseDuration(val)
}

func listParser(_ reflect.Type, val string, _ []string, kwParams map[string]string) (interface{}, error) {
	sep, ok := kwParams["sep"]
	if !ok {
		sep = ","
	}
	return strings.Split(val, sep), nil
}
