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
	sb_config_env_parser "github.com/SENERGY-Platform/go-service-base/config-hdl/env_parser"
	sb_config_types "github.com/SENERGY-Platform/go-service-base/config-hdl/types"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	"reflect"
	"strings"
)

var envTypeParser = []sb_config_hdl.EnvTypeParser{
	sb_config_types.SecretEnvTypeParser,
	sb_config_env_parser.DurationEnvTypeParser,
	logLevelEnvTypeParser,
	listEnvTypeParser,
}

func logLevelEnvTypeParser() (reflect.Type, sb_config_hdl.EnvParser) {
	return reflect.TypeOf(sb_logger.Off), logLevelEnvParser
}

func logLevelEnvParser(_ reflect.Type, val string, _ []string, _ map[string]string) (interface{}, error) {
	return sb_logger.ParseLevel(val)
}

func listEnvTypeParser() (reflect.Type, sb_config_hdl.EnvParser) {
	return reflect.TypeOf([]string{}), listEnvParser
}

func listEnvParser(_ reflect.Type, val string, _ []string, kwParams map[string]string) (interface{}, error) {
	sep, ok := kwParams["sep"]
	if !ok {
		sep = ","
	}
	return strings.Split(val, sep), nil
}
