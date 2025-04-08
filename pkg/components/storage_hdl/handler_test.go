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

package storage_hdl

import (
	"context"
	"github.com/SENERGY-Platform/go-service-base/struct-logger"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"os"
	"reflect"
	"testing"
)

func TestHandler(t *testing.T) {
	util.InitLogger(struct_logger.Config{}, os.Stderr, "", "")
	tmpDir := t.TempDir()
	hdl := New(tmpDir, "")
	t.Run("write 1", func(t *testing.T) {
		err := hdl.Write(context.Background(), "id-1", [][2]string{{"key", "/a"}}, []byte("test"))
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("write 2", func(t *testing.T) {
		err := hdl.Write(context.Background(), "id-2", [][2]string{{"key", "/b"}}, []byte("test"))
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("read", func(t *testing.T) {
		data, err := hdl.Read(context.Background(), "id-1")
		if err != nil {
			t.Error(err)
		}
		if string(data) != "test" {
			t.Errorf("expected 'test-1', got '%s'", string(data))
		}
	})
	t.Run("parallel", func(t *testing.T) {
		t.Run("read", func(t *testing.T) {
			t.Parallel()
			data, err := hdl.Read(context.Background(), "id-2")
			if err != nil {
				t.Error(err)
			}
			if string(data) != "test" {
				t.Errorf("expected 'test-2', got '%s'", string(data))
			}
		})
		t.Run("write 3", func(t *testing.T) {
			t.Parallel()
			err := hdl.Write(context.Background(), "id-3", [][2]string{{"key", "/c"}}, []byte("test"))
			if err != nil {
				t.Error(err)
			}
		})
		t.Run("delete", func(t *testing.T) {
			t.Parallel()
			err := hdl.Delete(context.Background(), "id-1")
			if err != nil {
				t.Error(err)
			}
		})
		t.Run("write 4", func(t *testing.T) {
			t.Parallel()
			err := hdl.Write(context.Background(), "id-4", [][2]string{{"key", "/d"}}, []byte("test"))
			if err != nil {
				t.Error(err)
			}
		})
		t.Run("list", func(t *testing.T) {
			t.Parallel()
			list, err := hdl.List(context.Background())
			if err != nil {
				t.Error(err)
			}
			if len(list) < 1 {
				t.Errorf("expected at least one items, got %d", len(list))
			}
			ok := false
			for _, item := range list {
				if item.ID == "id-2" {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("expected item with 'id-2' to be present")
			}
		})
	})
	t.Run("list", func(t *testing.T) {
		list, err := hdl.List(context.Background())
		if err != nil {
			t.Error(err)
		}
		if len(list) != 3 {
			t.Errorf("expected 3 items, got %d", len(list))
		}
	})
	t.Run("check items", func(t *testing.T) {
		t.Run("2", func(t *testing.T) {
			item, ok := hdl.items["id-2"]
			if !ok {
				t.Error("expected item 'id-2' to be present")
			}
			i := models.StorageData{
				ID:   "id-2",
				Args: [][2]string{{"key", "/b"}},
			}
			if !reflect.DeepEqual(item.StorageData, i) {
				t.Errorf("expected %v, got %v", i, item.StorageData)
			}
		})
		t.Run("3", func(t *testing.T) {
			item, ok := hdl.items["id-3"]
			if !ok {
				t.Error("expected item 'id-3' to be present")
			}
			i := models.StorageData{
				ID:   "id-3",
				Args: [][2]string{{"key", "/c"}},
			}
			if !reflect.DeepEqual(item.StorageData, i) {
				t.Errorf("expected %v, got %v", i, item.StorageData)
			}
		})
		t.Run("4", func(t *testing.T) {
			item, ok := hdl.items["id-4"]
			if !ok {
				t.Error("expected item 'id-4' to be present")
			}
			i := models.StorageData{
				ID:   "id-4",
				Args: [][2]string{{"key", "/d"}},
			}
			if !reflect.DeepEqual(item.StorageData, i) {
				t.Errorf("expected %v, got %v", i, item.StorageData)
			}
		})
	})
	t.Run("init", func(t *testing.T) {
		hdl2 := New(tmpDir, "")
		err := hdl2.Init(context.Background())
		if err != nil {
			t.Error(err)
		}
		if len(hdl2.items) != 3 {
			t.Errorf("expected 3 items, got %d", len(hdl2.items))
		}
		t.Run("check items", func(t *testing.T) {
			t.Run("2", func(t *testing.T) {
				item, ok := hdl2.items["id-2"]
				if !ok {
					t.Error("expected item 'id-2' to be present")
				}
				i := models.StorageData{
					ID:   "id-2",
					Args: [][2]string{{"key", "/b"}},
				}
				if !reflect.DeepEqual(item.StorageData, i) {
					t.Errorf("expected %v, got %v", i, item.StorageData)
				}
			})
			t.Run("3", func(t *testing.T) {
				item, ok := hdl2.items["id-3"]
				if !ok {
					t.Error("expected item 'id-3' to be present")
				}
				i := models.StorageData{
					ID:   "id-3",
					Args: [][2]string{{"key", "/c"}},
				}
				if !reflect.DeepEqual(item.StorageData, i) {
					t.Errorf("expected %v, got %v", i, item.StorageData)
				}
			})
			t.Run("4", func(t *testing.T) {
				item, ok := hdl2.items["id-4"]
				if !ok {
					t.Error("expected item 'id-4' to be present")
				}
				i := models.StorageData{
					ID:   "id-4",
					Args: [][2]string{{"key", "/d"}},
				}
				if !reflect.DeepEqual(item.StorageData, i) {
					t.Errorf("expected %v, got %v", i, item.StorageData)
				}
			})
		})
	})
	t.Run("error", func(t *testing.T) {
		t.Run("not found", func(t *testing.T) {
			t.Run("read", func(t *testing.T) {
				_, err := hdl.Read(context.Background(), "id-1")
				if err == nil {
					t.Error("expected error")
				}
			})
			t.Run("delete", func(t *testing.T) {
				err := hdl.Delete(context.Background(), "id-1")
				if err == nil {
					t.Error("expected error")
				}
			})
		})
		t.Run("path", func(t *testing.T) {
			hdl.dirPath = "does-not-exist"
			t.Run("write", func(t *testing.T) {
				err := hdl.Write(context.Background(), "", [][2]string{}, []byte(""))
				if err == nil {
					t.Error("expected error")
				}
			})
			t.Run("read", func(t *testing.T) {
				_, err := hdl.Read(context.Background(), "id-1")
				if err == nil {
					t.Error("expected error")
				}
			})
			t.Run("delete", func(t *testing.T) {
				err := hdl.Delete(context.Background(), "id-1")
				if err == nil {
					t.Error("expected error")
				}
			})
		})
	})
}
