package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetSwaggerDocsH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, SwaggerDocsPath, func(gc *gin.Context) {
		var userRoles []string
		if val := gc.GetHeader(HeaderUserRoles); val != "" {
			userRoles = strings.Split(val, ", ")
		}
		swaggerDocs, err := srv.GetSwaggerDocs(gc.Request.Context(), gc.Request.Header.Get("Authorization"), userRoles)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, swaggerDocs)
	}
}

func GetSrvInfoH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, InfoPath, func(gc *gin.Context) {
		gc.JSON(http.StatusOK, srv.GetSrvInfo(gc.Request.Context()))
	}
}

func GetHealthCheckH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, HealthCheckPath, func(gc *gin.Context) {
		err := srv.HealthCheck(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Status(http.StatusOK)
	}
}
