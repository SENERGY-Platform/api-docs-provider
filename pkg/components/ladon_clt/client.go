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

import (
	"context"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
)

type ClientItf interface {
	GetRoleAccessPolicy(ctx context.Context, role, path, method string) (bool, error)
	GetUserAccessPolicy(ctx context.Context, token string, pathMethodMap map[string][]string) (map[string][]string, error)
}

type Client struct {
	baseClient *base_client.Client
	baseUrl    string
}

func New(httpClient base_client.HTTPClient, baseUrl string) *Client {
	return &Client{
		baseClient: base_client.New(httpClient, customError, ""),
		baseUrl:    baseUrl,
	}
}

func customError(_ int, err error) error {
	return err
}
