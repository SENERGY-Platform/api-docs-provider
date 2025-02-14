package api

import (
	gin_mw "github.com/SENERGY-Platform/gin-middleware"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// New godoc
// @title Swagger-Docs-Provider API
// @version 0.5.3
// @description Provides swagger docs and storage management.
// @license.name Apache-2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /
func New(srv Service, staticHeader map[string]string) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	httpHandler := gin.New()
	httpHandler.Use(gin_mw.StaticHeaderHandler(staticHeader), requestid.New(requestid.WithCustomHeaderStrKey(HeaderRequestID)), gin_mw.LoggerHandler(util.Logger, []string{HealthCheckPath}, func(gc *gin.Context) string {
		return requestid.Get(gc)
	}), gin_mw.ErrorHandler(GetStatusCode, ", "), gin.Recovery())
	httpHandler.UseRawPath = true
	err := SetRoutes(httpHandler, srv)
	if err != nil {
		return nil, err
	}
	return httpHandler, nil
}
