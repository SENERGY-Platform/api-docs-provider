package api

import (
	"context"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetSwaggerDocsH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, SwaggerPath, func(gc *gin.Context) {
		var userRoles []string
		if val := gc.GetHeader(HeaderUserRoles); val != "" {
			userRoles = strings.Split(val, ", ")
		}
		swaggerDocs, err := srv.SwaggerDocs(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)), gc.Request.Header.Get(HeaderAuthorization), userRoles)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, swaggerDocs)
	}
}

func PatchRefreshDocsH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodPatch, DocsRefreshPath, func(gc *gin.Context) {
		err := srv.RefreshDocs(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Status(http.StatusOK)
	}
}

func GetDocsListH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, DocsListPath, func(gc *gin.Context) {
		list, err := srv.ListDocs(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, list)
	}
}

func GetSrvInfoH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, InfoPath, func(gc *gin.Context) {
		gc.JSON(http.StatusOK, srv.SrvInfo(gc.Request.Context()))
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
