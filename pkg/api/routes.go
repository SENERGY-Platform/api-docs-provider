package api

import (
	gin_mw "github.com/SENERGY-Platform/gin-middleware"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"github.com/gin-gonic/gin"
)

var routes = gin_mw.Routes[Service]{
	GetSwaggerDocsH,
	PatchRefreshDocsH,
	GetDocsListH,
	GetSrvInfoH,
	GetHealthCheckH,
}

func SetRoutes(e *gin.Engine, srv Service) error {
	err := routes.Set(srv, e, util.Logger)
	if err != nil {
		return err
	}
	return nil
}
