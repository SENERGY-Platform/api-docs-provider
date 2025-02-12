package storage_hdl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"github.com/google/uuid"
	"io"
	"io/fs"
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
}

func New(dirPath string) *Handler {
	return &Handler{
		dirPath: dirPath,
		items:   make(map[string]storageItem),
	}
}

func (h *Handler) Init(ctx context.Context) error {
	dirEntries, err := fs.ReadDir(os.DirFS(h.dirPath), ".")
	if err != nil {
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
				util.Logger.Errorf("storage: reading from '%s' failed: %v", se.dirName, err)
			}
			se.StorageData = data
			util.Logger.Debugf("storage: loaded '%s' from '%s'", se.ID, se.dirName)
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

func (h *Handler) Write(ctx context.Context, id string, extPaths []string, data []byte) error {
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
				util.Logger.Errorf("storage: %sremoving new dir '%s' of '%s' failed: %s", reqID, newDirName, id, e)
			}
		}
	}()
	item, ok := h.items[id]
	if !ok {
		item.ID = id
	}
	oldDirName := item.dirName
	item.dirName = newDirName
	item.ExtPaths = extPaths
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
			util.Logger.Errorf("storage: %sremoving old dir '%s' of '%s' failed: %s", reqID, oldDirName, id, e)
		}
	}
	util.Logger.Debugf("storage: %s'%s' written to '%s'", reqID, id, newDirName)
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
