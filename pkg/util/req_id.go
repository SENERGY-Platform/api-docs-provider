package util

import (
	"context"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
)

func GetReqID(ctx context.Context) string {
	val := ctx.Value(models.ContextRequestID)
	if val != nil {
		if str, ok := val.(string); ok {
			return str + " "
		}
	}
	return ""
}
