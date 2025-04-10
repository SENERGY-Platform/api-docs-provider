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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util/slog_attr"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/google/uuid"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"sync"
)

const (
	dataFileName = "data"
	docFileName  = "doc"
)

type Handler struct {
	dirPath string
	mu      sync.RWMutex
	items   map[string]storageItem
	logger  *slog.Logger
}

func New(dirPath, name string) *Handler {
	return &Handler{
		dirPath: dirPath,
		items:   make(map[string]storageItem),
		logger:  util.Logger.With(slog_attr.ComponentKey, name+"-storage-hdl"),
	}
}

func (h *Handler) Init(ctx context.Context) error {
	dirEntries, err := fs.ReadDir(os.DirFS(h.dirPath), ".")
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(h.dirPath, fs.ModePerm); err != nil {
				return err
			}
		}
		return err
	}
	for _, dirEntry := range dirEntries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if dirEntry.IsDir() {
			se := storageItem{dirName: dirEntry.Name()}
			data, err := readData(path.Join(h.dirPath, se.dirName, dataFileName))
			if err != nil {
				h.logger.Error("reading storage item failed", slog_attr.DirNameKey, se.dirName, attributes.ErrorKey, err)
			}
			se.StorageData = data
			h.logger.Debug("loaded storage item", slog_attr.IDKey, se.ID, slog_attr.DirNameKey, se.dirName)
			h.items[se.ID] = se
		}
	}
	return nil
}

func (h *Handler) List(_ context.Context) ([]models.StorageData, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var items []models.StorageData
	for _, item := range h.items {
		items = append(items, item.StorageData)
	}
	return items, nil
}

func (h *Handler) Write(ctx context.Context, id string, args [][2]string, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	var err error
	newDirName, err := genDirName()
	if err != nil {
		return models.NewInternalError(err)
	}
	err = os.Mkdir(path.Join(h.dirPath, newDirName), 0770)
	if err != nil {
		return models.NewInternalError(err)
	}
	reqID := util.GetReqID(ctx)
	defer func() {
		if err != nil {
			if e := os.RemoveAll(path.Join(h.dirPath, newDirName)); e != nil {
				h.logger.Error("removing new dir failed", slog_attr.DirNameKey, newDirName, slog_attr.IDKey, id, attributes.ErrorKey, e, slog_attr.RequestIDKey, reqID)
			}
		}
	}()
	item, ok := h.items[id]
	if !ok {
		item.ID = id
	}
	oldDirName := item.dirName
	item.dirName = newDirName
	item.Args = args
	dataFile, err := os.Create(path.Join(h.dirPath, newDirName, dataFileName))
	if err != nil {
		return models.NewInternalError(err)
	}
	defer dataFile.Close()
	err = json.NewEncoder(dataFile).Encode(item)
	if err != nil {
		return models.NewInternalError(err)
	}
	docFile, err := os.Create(path.Join(h.dirPath, newDirName, docFileName))
	if err != nil {
		return models.NewInternalError(err)
	}
	defer docFile.Close()
	n, err := docFile.ReadFrom(bytes.NewReader(data))
	if err != nil {
		return models.NewInternalError(err)
	}
	if n == 0 {
		err = models.NewInternalError(errors.New("0 bytes written"))
		return err
	}
	h.items[id] = item
	if oldDirName != "" {
		if e := os.RemoveAll(path.Join(h.dirPath, oldDirName)); e != nil {
			h.logger.Error("removing old dir failed", slog_attr.DirNameKey, oldDirName, slog_attr.IDKey, id, attributes.ErrorKey, e, slog_attr.RequestIDKey, reqID)
		}
	}
	h.logger.Debug("saved storage item", slog_attr.DirNameKey, newDirName, slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
	return nil
}

func (h *Handler) Read(_ context.Context, id string) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	item, ok := h.items[id]
	if !ok {
		return nil, models.NewNotFoundError(errors.New("not found"))
	}
	doc, err := readDoc(path.Join(h.dirPath, item.dirName, docFileName))
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	return doc, nil
}

func (h *Handler) Delete(_ context.Context, id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	item, ok := h.items[id]
	if !ok {
		return models.NewNotFoundError(errors.New("not found"))
	}
	err := os.RemoveAll(path.Join(h.dirPath, item.dirName))
	if err != nil {
		return models.NewInternalError(err)
	}
	delete(h.items, id)
	return nil
}

func genDirName() (string, error) {
	idObj, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return idObj.String(), nil
}

func readData(p string) (models.StorageData, error) {
	f, err := os.Open(p)
	if err != nil {
		return models.StorageData{}, err
	}
	defer f.Close()
	var data models.StorageData
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		return models.StorageData{}, err
	}
	return data, nil
}

func readDoc(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
