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

package kong_clt

type Route struct {
	//Tags                    interface{}  `json:"tags"`
	//StripPath               bool         `json:"strip_path"`
	//RegexPriority           int          `json:"regex_priority"`
	//Hosts                   interface{}  `json:"hosts"`
	Name string `json:"name"`
	ID   string `json:"id"`
	//PreserveHost            bool         `json:"preserve_host"`
	//CreatedAt               int          `json:"created_at"`
	//Sources                 interface{}  `json:"sources"`
	//Destinations            interface{}  `json:"destinations"`
	Paths []string `json:"paths"`
	//PathHandling            string       `json:"path_handling"`
	//Protocols               []string     `json:"protocols"`
	//Methods                 interface{}  `json:"methods"`
	//RequestBuffering        bool         `json:"request_buffering"`
	//ResponseBuffering       bool         `json:"response_buffering"`
	//Headers                 interface{}  `json:"headers"`
	//UpdatedAt               int          `json:"updated_at"`
	//Snis                    interface{}  `json:"snis"`
	//HttpsRedirectStatusCode int          `json:"https_redirect_status_code"`
	Service struct {
		ID string `json:"id"`
	} `json:"service"`
}

type Service struct {
	//Tags              interface{} `json:"tags"`
	//CaCertificates    interface{} `json:"ca_certificates"`
	//ClientCertificate interface{} `json:"client_certificate"`
	//Name           string      `json:"name"`
	//Path           interface{} `json:"path"`
	//ConnectTimeout int         `json:"connect_timeout"`
	//WriteTimeout   int         `json:"write_timeout"`
	//TlsVerify      interface{} `json:"tls_verify"`
	//TlsVerifyDepth interface{} `json:"tls_verify_depth"`
	//UpdatedAt      int         `json:"updated_at"`
	Host     string `json:"host"`
	Protocol string `json:"protocol"`
	//Retries        int         `json:"retries"`
	//Enabled        bool        `json:"enabled"`
	ID string `json:"id"`
	//CreatedAt      int         `json:"created_at"`
	//ReadTimeout    int         `json:"read_timeout"`
	Port int `json:"port"`
}

type routes struct {
	Data []Route `json:"data"`
}

type services struct {
	Data []Service `json:"data"`
}
