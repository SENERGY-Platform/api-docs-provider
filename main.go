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

package main

import (
	"context"
	"fmt"
	lib_models "github.com/SENERGY-Platform/api-docs-provider/lib/models"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/api"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/components/discovery_hdl"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/components/doc_clt"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/components/kong_clt"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/components/ladon_clt"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/components/storage_hdl"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/config"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/service"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/service/asyncapi_srv"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/service/swagger_srv"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/util/slog_attr"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	srv_info_hdl "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl"
	sb_util "github.com/SENERGY-Platform/mgw-go-service-base/util"
	"net/http"
	"os"
	"sync"
	"syscall"
)

var version string

func main() {
	srvInfoHdl := srv_info_hdl.New("api-docs-provider", version)

	ec := 0
	defer func() {
		os.Exit(ec)
	}()

	util.ParseFlags()

	cfg, err := config.New(util.Flags.ConfPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		ec = 1
		return
	}

	util.InitLogger(cfg.Logger, os.Stderr, "github.com/SENERGY-Platform", srvInfoHdl.GetName())
	discovery_hdl.InitLogger()
	swagger_srv.InitLogger()
	asyncapi_srv.InitLogger()

	util.Logger.Info("starting service", slog_attr.VersionKey, srvInfoHdl.GetVersion())

	util.Logger.Debug(sb_util.ToJsonStr(cfg))

	swaggerStgHdl := storage_hdl.New(cfg.Storage.SwaggerDataPath, "swagger")
	kongClt := kong_clt.New(&http.Client{Transport: http.DefaultTransport}, cfg.Discovery.Kong.BaseURL, cfg.Discovery.Kong.User, cfg.Discovery.Kong.Password.Value())
	discoveryHdl := discovery_hdl.New(kongClt, cfg.HttpTimeout, cfg.Discovery.HostBlacklist)
	docClt := doc_clt.New(&http.Client{Transport: http.DefaultTransport}, cfg.Procurement.SwaggerDocPath)
	ladonClt := ladon_clt.New(&http.Client{Transport: http.DefaultTransport}, cfg.Filter.LadonBaseUrl)
	swaggerSrv := swagger_srv.New(swaggerStgHdl, discoveryHdl, docClt, ladonClt, cfg.HttpTimeout, cfg.ApiGateway, cfg.Filter.AdminRoleName)

	asyncapiStgHdl := storage_hdl.New(cfg.Storage.AsyncapiDataPath, "asyncapi")
	asyncapiSrv := asyncapi_srv.New(asyncapiStgHdl)

	srv := service.New(swaggerSrv, asyncapiSrv, srvInfoHdl)

	httpHandler, err := api.New(srv, map[string]string{
		lib_models.HeaderApiVer:  srvInfoHdl.GetVersion(),
		lib_models.HeaderSrvName: srvInfoHdl.GetName(),
	}, cfg.HttpAccessLog)
	if err != nil {
		util.Logger.Error("creating http engine failed", attributes.ErrorKey, err)
		ec = 1
		return
	}

	httpServer := util.NewServer(httpHandler, cfg.ServerPort)

	ctx, cf := context.WithCancel(context.Background())

	go func() {
		util.WaitForSignal(ctx, syscall.SIGINT, syscall.SIGTERM)
		cf()
	}()

	if err = swaggerStgHdl.Init(ctx); err != nil {
		util.Logger.Error("initializing swagger storage handler failed", attributes.ErrorKey, err)
		ec = 1
		return
	}

	if err = asyncapiStgHdl.Init(ctx); err != nil {
		util.Logger.Error("initializing asyncapi storage handler failed", attributes.ErrorKey, err)
		ec = 1
		return
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := swaggerSrv.SwaggerPeriodicProcurement(ctx, cfg.Procurement.Interval); err != nil {
			util.Logger.Error("periodic procurement failed", attributes.ErrorKey, err)
			ec = 1
		}
		cf()
	}()

	go func() {
		if err := util.StartServer(httpServer); err != nil {
			util.Logger.Error("starting server failed", attributes.ErrorKey, err)
			ec = 1
		}
		cf()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := util.StopServer(ctx, httpServer); err != nil {
			util.Logger.Error("stopping server failed", attributes.ErrorKey, err)
			ec = 1
		}
	}()

	wg.Wait()
}
