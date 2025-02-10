package api

import (
	"context"
	"encoding/json"
	srv_info_lib "github.com/SENERGY-Platform/go-service-base/srv-info-hdl/lib"
)

type Service interface {
	GetSwaggerDocs(ctx context.Context, userRoles []string) ([]map[string]json.RawMessage, error)
	HealthCheck(ctx context.Context) error
	GetSrvInfo(_ context.Context) srv_info_lib.SrvInfo
}
