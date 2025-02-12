package api

import (
	"context"
	"encoding/json"
	srv_info_lib "github.com/SENERGY-Platform/go-service-base/srv-info-hdl/lib"
)

type Service interface {
	SwaggerDocs(ctx context.Context, userToken string, userRoles []string) ([]map[string]json.RawMessage, error)
	RefreshDocs(ctx context.Context) error
	HealthCheck(ctx context.Context) error
	SrvInfo(ctx context.Context) srv_info_lib.SrvInfo
}
