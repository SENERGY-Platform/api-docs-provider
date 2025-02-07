package storage_hdl

import "github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"

type storageItem struct {
	models.StorageData
	dirName string
}
