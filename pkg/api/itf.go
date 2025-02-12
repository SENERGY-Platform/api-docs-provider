package api

import (
	"context"
	"encoding/json"
	srv_info_lib "github.com/SENERGY-Platform/go-service-base/srv-info-hdl/lib"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
)

type Service interface {
	SwaggerDocs(ctx context.Context, userToken string, userRoles []string) ([]map[string]json.RawMessage, error)
	RefreshStorage(ctx context.Context) error
	ListStorage(ctx context.Context) ([]models.StorageData, error)
	HealthCheck(ctx context.Context) error
	SrvInfo(ctx context.Context) srv_info_lib.SrvInfo
}
