package service

import (
	"context"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
)

type DiscoveryHandler interface {
	GetServices(ctx context.Context) (map[string]models.Service, error)
}

type StorageHandler interface {
	List(ctx context.Context) ([]models.StorageData, error)
	Write(ctx context.Context, id string, extPaths []string, data []byte) error
	Read(ctx context.Context, id string) ([]byte, error)
	Delete(ctx context.Context, id string) error
}
