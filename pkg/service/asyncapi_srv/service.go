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

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/models"
	srv_util "github.com/SENERGY-Platform/api-docs-provider/pkg/service/util"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util/slog_attr"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"sync"
)

const (
	titleArgKey   = "title"
	versionArgKey = "version"
)

type Service struct {
	storageHdl StorageHandler
}

func New(storageHdl StorageHandler) *Service {
	return &Service{
		storageHdl: storageHdl,
	}
}

func (s *Service) AsyncapiGetDocs(ctx context.Context) ([]json.RawMessage, error) {
	items, err := s.storageHdl.List(ctx)
	if err != nil {
		return []json.RawMessage{}, models.NewInternalError(err)
	}
	reqID := util.GetReqID(ctx)
	var docs []json.RawMessage
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for _, item := range items {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			logger.Debug("reading doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
			rawDoc, err := s.storageHdl.Read(ctx, id)
			if err != nil {
				logger.Error("reading doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
				return
			}
			var doc json.RawMessage
			if err = json.Unmarshal(rawDoc, &doc); err != nil {
				logger.Error("transforming doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err.Error(), slog_attr.RequestIDKey, reqID)
				return
			}
			mu.Lock()
			docs = append(docs, doc)
			mu.Unlock()
			logger.Debug("appended doc", slog_attr.IDKey, id, slog_attr.RequestIDKey, reqID)
		}(item.ID)
	}
	wg.Wait()
	return docs, nil
}

func (s *Service) AsyncapiPutDoc(ctx context.Context, id string, data []byte) error {
	reqID := util.GetReqID(ctx)
	if err := validateDoc(data); err != nil {
		logger.Error("validating doc failed", slog_attr.IDKey, id, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return models.NewInvalidInputError(err)
	}
	title, version, err := getAsyncapiInfo(data)
	if err != nil {
		logger.Error("extracting info failed", slog_attr.IDKey, id, attributes.ErrorKey, err, slog_attr.RequestIDKey, reqID)
		return models.NewInternalError(err)
	}
	return s.storageHdl.Write(ctx, id, [][2]string{
		{titleArgKey, title},
		{versionArgKey, version},
	}, data)
}

func (s *Service) AsyncapiDeleteDoc(ctx context.Context, id string) error {
	return s.storageHdl.Delete(ctx, id)
}

func (s *Service) AsyncapiListStorage(ctx context.Context) ([]models.AsyncapiItem, error) {
	storageItems, err := s.storageHdl.List(ctx)
	if err != nil {
		return nil, err
	}
	var asyncapiItems []models.AsyncapiItem
	for _, storageItem := range storageItems {
		asyncapiItems = append(asyncapiItems, newAsyncapiItem(storageItem))
	}
	return asyncapiItems, nil
}

func newAsyncapiItem(sd models.StorageData) models.AsyncapiItem {
	ai := models.AsyncapiItem{ID: sd.ID}
	for _, arg := range sd.Args {
		switch arg[0] {
		case titleArgKey:
			ai.Title = arg[1]
		case versionArgKey:
			ai.Version = arg[1]
		}
	}
	return ai
}

func validateDoc(doc []byte) error {
	var tmp map[string]json.RawMessage
	if err := json.Unmarshal(doc, &tmp); err != nil {
		return err
	}
	if !srv_util.CheckForKeys(tmp, asyncapiV2Keys) && !srv_util.CheckForKeys(tmp, asyncapiV3Keys) {
		return errors.New("missing required keys")
	}
	return nil
}

func getAsyncapiInfo(doc []byte) (string, string, error) {
	var info asyncapiInfo
	if err := json.Unmarshal(doc, &info); err != nil {
		return "", "", err
	}
	return info.Info.Title, info.Info.Version, nil
}
