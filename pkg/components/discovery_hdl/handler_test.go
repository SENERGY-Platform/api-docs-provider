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

package discovery_hdl

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/kong_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/config"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"os"
	"reflect"
	"testing"
)

func TestHandler_GetServices(t *testing.T) {
	mockClt := &mockClient{
		Routes: []kong_clt.Route{
			{
				Name:  "route-a",
				ID:    "r1",
				Paths: []string{"/a/a", "/a/b"},
				Service: struct {
					ID string `json:"id"`
				}{ID: "s1"},
			},
			{
				Name:  "route-b",
				ID:    "r2",
				Paths: []string{"/c"},
				Service: struct {
					ID string `json:"id"`
				}{ID: "s2"},
			},
			{
				Name:  "route-c",
				ID:    "r3",
				Paths: []string{"/d"},
				Service: struct {
					ID string `json:"id"`
				}{ID: "s2"},
			},
			{
				Name:  "route-d",
				ID:    "r4",
				Paths: []string{"/e"},
				Service: struct {
					ID string `json:"id"`
				}{ID: "s3"},
			},
		},
		Services: []kong_clt.Service{
			{
				Host:     "api.srv-a",
				Protocol: "http",
				ID:       "s1",
				Port:     8000,
			},
			{
				Host:     "api.srv-b",
				Protocol: "https",
				ID:       "s2",
				Port:     8080,
			},
			{
				Host:     "api.srv-c",
				Protocol: "https",
				ID:       "s3",
				Port:     80,
			},
		},
	}
	util.InitLogger(config.LoggerConfig{}, os.Stderr, "", "")
	hdl := New(mockClt, 0, []string{"api.srv-c"})
	a := map[string]models.Service{
		"api.srv-a8000": {
			ID:       "api.srv-a8000",
			Host:     "api.srv-a",
			Port:     8000,
			Protocol: "http",
			ExtPaths: []string{"/a/a", "/a/b"},
		},
		"api.srv-b8080": {
			ID:       "api.srv-b8080",
			Host:     "api.srv-b",
			Port:     8080,
			Protocol: "https",
			ExtPaths: []string{"/c", "/d"},
		},
	}
	b, err := hdl.GetServices(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expected: %v, got: %v", a, b)
	}
	t.Run("error", func(t *testing.T) {
		t.Run("get routes", func(t *testing.T) {
			mockClt.GetRoutesErr = errors.New("error")
			_, err = hdl.GetServices(context.Background())
			if err == nil {
				t.Error("expected error")
			}
			mockClt.GetRoutesErr = nil
		})
		t.Run("get services", func(t *testing.T) {
			mockClt.GetServicesErr = errors.New("error")
			_, err = hdl.GetServices(context.Background())
			if err == nil {
				t.Error("expected error")
			}
			mockClt.GetServicesErr = nil
		})
	})
}

func Test_getKongSrvMap(t *testing.T) {
	a := map[string]kong_clt.Service{
		"1": {
			Host:     "api.a",
			Protocol: "http",
			ID:       "1",
			Port:     8000,
		},
		"2": {
			Host:     "api.b",
			Protocol: "https",
			ID:       "2",
			Port:     8080,
		},
	}
	b := getKongSrvMap([]kong_clt.Service{
		{
			Host:     "api.a",
			Protocol: "http",
			ID:       "1",
			Port:     8000,
		},
		{
			Host:     "api.b",
			Protocol: "https",
			ID:       "2",
			Port:     8080,
		},
	})
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expected: %v, got: %v", a, b)
	}
}

type mockClient struct {
	Routes         []kong_clt.Route
	Services       []kong_clt.Service
	Err            error
	GetRoutesErr   error
	GetServicesErr error
	GetRoutesC     int
	GetServicesC   int
}

func (m *mockClient) GetRoutes(_ context.Context) ([]kong_clt.Route, error) {
	m.GetRoutesC++
	if m.Err != nil {
		return nil, m.Err
	}
	if m.GetRoutesErr != nil {
		return nil, m.GetRoutesErr
	}
	return m.Routes, nil
}

func (m *mockClient) GetServices(_ context.Context) ([]kong_clt.Service, error) {
	m.GetServicesC++
	if m.Err != nil {
		return nil, m.Err
	}
	if m.GetServicesErr != nil {
		return nil, m.GetServicesErr
	}
	return m.Services, nil
}
