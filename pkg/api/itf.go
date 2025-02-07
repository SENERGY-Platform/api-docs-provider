package api

import (
	"context"
	"encoding/json"
)

type Service interface {
	GetSwaggerDocs(ctx context.Context, userRoles []string) ([]map[string]json.RawMessage, error)
}
