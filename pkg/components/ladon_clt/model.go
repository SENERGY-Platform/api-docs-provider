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

package ladon_clt

type roleAccessRequest struct {
	Resource string `json:"resource"` // resource that access is requested to
	Action   string `json:"action"`   // action that is requested on the resource
	Subject  string `json:"subject"`  // subject that is requesting access
}

type roleAccessResponse struct {
	Result bool   `json:"result"`
	Error  string `json:"error"`
}

type userAccessRequest struct {
	Method      string `json:"method"`
	Endpoint    string `json:"endpoint"`
	orgMethod   string
	orgEndpoint string
}

type userAccessResponse struct {
	Allowed []bool `json:"allowed"`
}
