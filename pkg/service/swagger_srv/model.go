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

package swagger_srv

import (
	"encoding/json"
)

const (
	swaggerKey            = "swagger"
	swaggerInfoKey        = "info"
	swaggerOpenApiKey     = "openapi"
	swaggerHostKey        = "host"
	swaggerBasePathKey    = "basePath"
	swaggerSchemesKey     = "schemes"
	swaggerPathsKey       = "paths"
	swaggerDefinitionsKey = "definitions"
)

var swaggerV2Keys = []string{
	swaggerKey,
	swaggerInfoKey,
	swaggerPathsKey,
}

var swaggerV3Keys = []string{
	swaggerInfoKey,
	swaggerOpenApiKey,
	swaggerPathsKey,
}

type docWrapper struct {
	basePath string
	doc      map[string]json.RawMessage
}

type swaggerDoc struct {
	Info swaggerInfo `json:"info"`
}

type swaggerInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}
