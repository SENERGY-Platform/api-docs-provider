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

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/config"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestService_RefreshStorage(t *testing.T) {
	validDoc, err := json.Marshal(map[string]json.RawMessage{
		"swagger": nil,
		"info":    nil,
		"paths":   nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	storageHdl := &storageHdlMock{
		Items: map[string]struct {
			models.StorageData
			data []byte
		}{
			"ph0": {
				StorageData: models.StorageData{
					ID:       "ph0",
					ExtPaths: []string{"/x"},
				},
				data: validDoc,
			},
			"ph9": {
				StorageData: models.StorageData{
					ID:       "ph9",
					ExtPaths: []string{"/t"},
				},
				data: validDoc,
			},
		},
	}
	docClt := &docCltMock{
		Docs: map[string][]byte{
			"ph0": validDoc,
			"ph1": validDoc,
			"ph2": []byte("test"),
		},
	}
	discoveryHdl := &discoveryHdlMock{
		Services: map[string]models.Service{
			"ph0": {
				ID:       "ph0",
				Host:     "h",
				Port:     0,
				Protocol: "p",
				ExtPaths: []string{"/t"},
			},
			"ph1": {
				ID:       "ph1",
				Host:     "h",
				Port:     1,
				Protocol: "p",
				ExtPaths: []string{"/t"},
			},
			"ph2": {
				ID:       "ph2",
				Host:     "h",
				Port:     2,
				Protocol: "p",
				ExtPaths: []string{"/t"},
			},
			"ph3": {
				ID:       "ph3",
				Host:     "h",
				Port:     3,
				Protocol: "p",
				ExtPaths: []string{"/t"},
			},
			"ph4": {
				ID:       "ph4",
				Host:     "h",
				Port:     4,
				Protocol: "p",
			},
		},
	}
	util.InitLogger(config.LoggerConfig{}, os.Stderr, "", "")
	InitLogger()
	srv := New(storageHdl, discoveryHdl, nil, docClt, nil, 0, "", "")
	err = srv.RefreshStorage(context.Background())
	if err != nil {
		t.Error(err)
	}
	a := map[string]struct {
		models.StorageData
		data []byte
	}{
		"ph0": {
			StorageData: models.StorageData{
				ID:       "ph0",
				ExtPaths: []string{"/t"},
			},
			data: validDoc,
		},
		"ph1": {
			StorageData: models.StorageData{
				ID:       "ph1",
				ExtPaths: []string{"/t"},
			},
			data: validDoc,
		},
	}
	if !reflect.DeepEqual(a, storageHdl.Items) {
		t.Errorf("expected %v, got %v", a, storageHdl.Items)
	}
}

func Test_validateDoc(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		t.Run("v2", func(t *testing.T) {
			swV2 := map[string]json.RawMessage{
				"swagger": nil,
				"info":    nil,
				"paths":   nil,
			}
			b, err := json.Marshal(swV2)
			if err != nil {
				t.Fatal(err)
			}
			if err = validateDoc(b); err != nil {
				t.Error("unexpected error:", err)
			}
		})
		t.Run("v3", func(t *testing.T) {
			swV3 := map[string]any{
				"info":    nil,
				"openapi": nil,
				"paths":   nil,
			}
			b, err := json.Marshal(swV3)
			if err != nil {
				t.Fatal(err)
			}
			if err = validateDoc(b); err != nil {
				t.Error("unexpected error:", err)
			}
		})
	})
	t.Run("invalid", func(t *testing.T) {
		t.Run("missing keys", func(t *testing.T) {
			doc := map[string]json.RawMessage{
				"info":   nil,
				"status": nil,
			}
			b, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if err = validateDoc(b); err == nil {
				t.Error("expected error")
			}
		})
		t.Run("not json", func(t *testing.T) {
			if err := validateDoc([]byte("test")); err == nil {
				t.Error("expected error")
			}
		})
	})
}

func TestService_cleanOldServices(t *testing.T) {
	sHdl := &storageHdlMock{
		Items: map[string]struct {
			models.StorageData
			data []byte
		}{
			"id-1": {
				StorageData: models.StorageData{
					ID:       "id-1",
					ExtPaths: nil,
				},
			},
			"id-2": {
				StorageData: models.StorageData{
					ID:       "id-2",
					ExtPaths: nil,
				},
			},
		},
	}
	srv := New(sHdl, nil, nil, nil, nil, 0, "", "")
	err := srv.cleanOldServices(context.Background(), map[string]models.Service{
		"id-2": {
			ID: "id-2",
		},
	})
	if err != nil {
		t.Error(err)
	}
	if len(sHdl.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(sHdl.Items))
	}
	if _, ok := sHdl.Items["id-2"]; !ok {
		t.Error("expected 'id-2'")
	}
}

type docCltMock struct {
	Docs map[string][]byte
	Err  error
}

func (m *docCltMock) GetDoc(_ context.Context, protocol, host string, port int) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	b, ok := m.Docs[fmt.Sprintf("%s%s%d", protocol, host, port)]
	if !ok {
		return nil, errors.New("not found")
	}
	return b, nil
}

type discoveryHdlMock struct {
	Services map[string]models.Service
	Err      error
}

func (m *discoveryHdlMock) GetServices(_ context.Context) (map[string]models.Service, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Services, nil
}

type storageHdlMock struct {
	Items map[string]struct {
		models.StorageData
		data []byte
	}
	Err       error
	WriteErr  error
	ReadErr   error
	DeleteErr error
	mu        sync.RWMutex
}

func (m *storageHdlMock) List(_ context.Context) ([]models.StorageData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Err != nil {
		return nil, m.Err
	}
	var list []models.StorageData
	for _, item := range m.Items {
		list = append(list, item.StorageData)
	}
	return list, nil
}

func (m *storageHdlMock) Write(_ context.Context, id string, extPaths []string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return m.Err
	}
	if m.WriteErr != nil {
		return m.WriteErr
	}
	m.Items[id] = struct {
		models.StorageData
		data []byte
	}{
		StorageData: models.StorageData{
			ID:       id,
			ExtPaths: extPaths,
		},
		data: data,
	}
	return nil
}

func (m *storageHdlMock) Read(_ context.Context, id string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Err != nil {
		return nil, m.Err
	}
	if m.ReadErr != nil {
		return nil, m.ReadErr
	}
	item, ok := m.Items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return item.data, nil
}

func (m *storageHdlMock) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return m.Err
	}
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	_, ok := m.Items[id]
	if !ok {
		return errors.New("not found")
	}
	delete(m.Items, id)
	return nil
}
