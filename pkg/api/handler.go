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
	"errors"
	lib_models "github.com/SENERGY-Platform/api-docs-provider/lib/models"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util"
	_ "github.com/SENERGY-Platform/go-service-base/srv-info-hdl"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strings"
)

// getSwaggerGetDocsOldH godoc
// @Summary Get docs
// @Description Get all swagger docs.
// @Tags Swagger
// @Produce	json
// @Param Authorization header string false "jwt token"
// @Param X-User-Roles header string false "user roles"
// @Success	200 {array} object "list of swagger docs"
// @Failure	500 {string} string "error message"
// @Router /swagger [get]
// @Deprecated
func getSwaggerGetDocsOldH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/swagger", func(gc *gin.Context) {
		var userRoles []string
		if val := gc.GetHeader(HeaderUserRoles); val != "" {
			userRoles = strings.Split(val, ", ")
		}
		docs, err := srv.SwaggerGetDocs(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)), gc.Request.Header.Get(HeaderAuthorization), userRoles)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, docs)
	}
}

// getSwaggerGetDocsH godoc
// @Summary Get docs
// @Description Get all swagger docs.
// @Tags Swagger
// @Produce	json
// @Param Authorization header string false "jwt token"
// @Param X-User-Roles header string false "user roles"
// @Success	200 {array} object "list of swagger docs"
// @Failure	500 {string} string "error message"
// @Router /docs/swagger [get]
func getSwaggerGetDocsH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/docs/swagger", func(gc *gin.Context) {
		var userRoles []string
		if val := gc.GetHeader(HeaderUserRoles); val != "" {
			userRoles = strings.Split(val, ", ")
		}
		docs, err := srv.SwaggerGetDocs(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)), gc.Request.Header.Get(HeaderAuthorization), userRoles)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, docs)
	}
}

// getSwaggerGetDocH godoc
// @Summary Get doc
// @Description Get a swagger doc.
// @Tags Swagger
// @Produce	json
// @Param Authorization header string false "jwt token"
// @Param X-User-Roles header string false "user roles"
// @Param id path string true "doc id"
// @Success	200 {object} object "swagger doc"
// @Failure	403 {string} string "error message"
// @Failure	404 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /docs/swagger/{id} [get]
func getSwaggerGetDocH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/docs/swagger/:id", func(gc *gin.Context) {
		var userRoles []string
		if val := gc.GetHeader(HeaderUserRoles); val != "" {
			userRoles = strings.Split(val, ", ")
		}
		doc, err := srv.SwaggerGetDoc(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)), gc.Param("id"), gc.Request.Header.Get(HeaderAuthorization), userRoles)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Data(http.StatusOK, gin.MIMEJSON, doc)
	}
}

// patchSwaggerRefreshDocsH godoc
// @Summary Refresh storage
// @Description Trigger swagger docs refresh.
// @Tags Swagger
// @Success	200
// @Failure	500 {string} string "error message"
// @Router /storage-refresh/swagger [patch]
func patchSwaggerRefreshDocsH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/storage-refresh/swagger", func(gc *gin.Context) {
		err := srv.SwaggerRefreshDocs(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Status(http.StatusOK)
	}
}

// getSwaggerListStorageH godoc
// @Summary List storage
// @Description Get meta information of all stored items.
// @Tags Swagger
// @Produce	json
// @Success	200 {array} models.SwaggerItem "stored items"
// @Failure	500 {string} string "error message"
// @Router /storage/swagger [get]
func getSwaggerListStorageH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/storage/swagger", func(gc *gin.Context) {
		var userRoles []string
		if val := gc.GetHeader(HeaderUserRoles); val != "" {
			userRoles = strings.Split(val, ", ")
		}
		items, err := srv.SwaggerListStorage(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)), gc.Request.Header.Get(HeaderAuthorization), userRoles)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, items)
	}
}

// getAsyncapiGetDocsH godoc
// @Summary Get docs
// @Description Get all asyncapi docs.
// @Tags AsyncAPI
// @Produce	json
// @Param Authorization header string false "jwt token"
// @Success	200 {array} object "list of asyncapi docs"
// @Failure	500 {string} string "error message"
// @Router /docs/asyncapi [get]
func getAsyncapiGetDocsH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/docs/asyncapi", func(gc *gin.Context) {
		docs, err := srv.AsyncapiGetDocs(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, docs)
	}
}

// getAsyncapiGetDocH godoc
// @Summary Get doc
// @Description Get an asyncapi doc.
// @Tags AsyncAPI
// @Produce	json
// @Param Authorization header string false "jwt token"
// @Param id path string true "doc id"
// @Success	200 {object} object "asyncapi doc"
// @Failure	404 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /docs/asyncapi/{id} [get]
func getAsyncapiGetDocH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/docs/asyncapi/:id", func(gc *gin.Context) {
		doc, err := srv.AsyncapiGetDoc(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)), gc.Param("id"))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Data(http.StatusOK, gin.MIMEJSON, doc)
	}
}

// getAsyncapiListStorage godoc
// @Summary List storage
// @Description Get meta information of all stored items.
// @Tags AsyncAPI
// @Accept json
// @Param Authorization header string false "jwt token"
// @Success	200 {array} models.AsyncapiItem "stored items"
// @Failure	500 {string} string "error message"
// @Router /storage/asyncapi [get]
func getAsyncapiListStorage(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/storage/asyncapi", func(gc *gin.Context) {
		items, err := srv.AsyncapiListStorage(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, items)
	}
}

// putAsyncapiPutDocH godoc
// @Summary Store doc
// @Description Store an asyncapi doc.
// @Tags AsyncAPI
// @Accept octet-stream
// @Param Authorization header string false "jwt token"
// @Param id path string true "doc id"
// @Param data body string true "doc"
// @Success	200
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /storage/asyncapi/{id} [put]
func putAsyncapiPutDocH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/storage/asyncapi/:id", func(gc *gin.Context) {
		id := gc.Param("id")
		if id == "" {
			_ = gc.Error(lib_models.NewInvalidInputError(errors.New("id is required")))
			return
		}
		data, err := io.ReadAll(gc.Request.Body)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		err = srv.AsyncapiPutDoc(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)), id, data)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Status(http.StatusOK)
	}
}

// deleteAsyncapiDeleteDocH godoc
// @Summary Delete doc
// @Description Remove an asyncapi doc.
// @Tags AsyncAPI
// @Param Authorization header string false "jwt token"
// @Param id path string true "doc id"
// @Success	200
// @Failure	400 {string} string "error message"
// @Failure	404 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /storage/asyncapi/{id} [delete]
func deleteAsyncapiDeleteDocH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/storage/asyncapi/:id", func(gc *gin.Context) {
		err := srv.AsyncapiDeleteDoc(context.WithValue(gc.Request.Context(), models.ContextRequestID, requestid.Get(gc)), gc.Param("id"))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Status(http.StatusOK)
	}
}

// getInfoH godoc
// @Summary Get service info
// @Description	Get basic service and runtime information.
// @Tags Info
// @Produce	json
// @Success	200 {object} srv_info_hdl.ServiceInfo "info"
// @Failure	500 {string} string "error message"
// @Router /info [get]
func getInfoH(srv Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/info", func(gc *gin.Context) {
		gc.JSON(http.StatusOK, srv.ServiceInfo())
	}
}

func getHealthCheckH(_ Service) (string, string, gin.HandlerFunc) {
	return http.MethodGet, HealthCheckPath, func(gc *gin.Context) {
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
