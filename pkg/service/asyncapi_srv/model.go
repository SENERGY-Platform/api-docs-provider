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

package asyncapi_srv

const (
	asyncapiKey         = "asyncapi"
	asyncapiInfoKey     = "info"
	asyncapiChannelsKey = "channels"
)

var asyncapiV2Keys = []string{
	asyncapiKey,
	asyncapiInfoKey,
	asyncapiChannelsKey,
}

var asyncapiV3Keys = []string{
	asyncapiKey,
	asyncapiInfoKey,
}

type asyncapiDoc struct {
	Info asyncapiInfo `json:"info"`
}

type asyncapiInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}
