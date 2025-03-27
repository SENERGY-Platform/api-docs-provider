/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"context"
	"github.com/SENERGY-Platform/go-service-base/structured-logger/attributes"
	_ "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl/lib"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

// getSwaggerDocsH godoc
// @Summary Get swagger docs
// @Description Get all swagger documents.
// @Tags Swagger
// @Produce	json
// @Param Authorization header string false "jwt token"
// @Param X-User-Roles header string false "user roles"
// @Success	200 {array} object "list of swagger docs"
// @Failure	500 {string} string "error message"
// @Router /swagger [get]
func getSwaggerDocsH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/swagger", func(gc *gin.Context) {
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

// patchStorageRefreshH godoc
// @Summary Refresh storage
// @Description Trigger swagger docs refresh.
// @Tags Storage
// @Success	200
// @Failure	500 {string} string "error message"
// @Router /storage/refresh [patch]
func patchStorageRefreshH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/storage/refresh", func(gc *gin.Context) {
		err := srv.RefreshStorage(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Status(http.StatusOK)
	}
}

// getStorageListH godoc
// @Summary List storage
// @Description Get meta information of all stored items.
// @Tags Storage
// @Produce	json
// @Success	200 {array} models.StorageData "stored items"
// @Failure	500 {string} string "error message"
// @Router /storage/list [get]
func getStorageListH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/storage/list", func(gc *gin.Context) {
		list, err := srv.ListStorage(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, list)
	}
}

// getInfoH godoc
// @Summary Get service info
// @Description	Get basic service and runtime information.
// @Tags Info
// @Produce	json
// @Success	200 {object} lib.SrvInfo "info"
// @Failure	500 {string} string "error message"
// @Router /info [get]
func getInfoH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/info", func(gc *gin.Context) {
		gc.JSON(http.StatusOK, srv.SrvInfo(gc.Request.Context()))
	}
}

func getHealthCheckH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, HealthCheckPath, func(gc *gin.Context) {
		err := srv.HealthCheck(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Status(http.StatusOK)
	}
}

func getSwaggerDocH(_ Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/doc", func(gc *gin.Context) {
		if _, err := os.Stat("docs/swagger.json"); err != nil {
			_ = gc.Error(err)
			util.Logger.Error("reading swagger file failed", attributes.ErrorKey, err)
			return
		}
		gc.Header("Content-Type", gin.MIMEJSON)
		gc.File("docs/swagger.json")
	}
}
