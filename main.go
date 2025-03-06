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
	"errors"
	"fmt"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	srv_info_hdl "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl"
	sb_util "github.com/SENERGY-Platform/mgw-go-service-base/util"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/api"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/discovery_hdl"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/doc_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/kong_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/ladon_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/storage_hdl"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/config"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/service"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"net/http"
	"os"
	"sync"
	"syscall"
)

var version string

func main() {
	srvInfoHdl := srv_info_hdl.New("swagger-docs-provider", version)

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

	logFile, err := util.InitLogger(cfg.Logger)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		var logFileError *sb_logger.LogFileError
		if errors.As(err, &logFileError) {
			ec = 1
			return
		}
	}
	if logFile != nil {
		defer logFile.Close()
	}

	util.Logger.Printf("%s %s", srvInfoHdl.GetName(), srvInfoHdl.GetVersion())

	util.Logger.Debugf("config: %s", sb_util.ToJsonStr(cfg))

	storageHdl := storage_hdl.New(cfg.WorkdirPath)

	kongClt := kong_clt.New(&http.Client{Transport: http.DefaultTransport}, cfg.Discovery.Kong.BaseURL, cfg.Discovery.Kong.User, cfg.Discovery.Kong.Password.Value())

	discoveryHdl := discovery_hdl.New(kongClt, cfg.HttpTimeout, cfg.Discovery.HostBlacklist)

	docClt := doc_clt.New(&http.Client{Transport: http.DefaultTransport}, cfg.Procurement.SwaggerDocPath)

	ladonClt := ladon_clt.New(&http.Client{Transport: http.DefaultTransport}, cfg.Filter.LadonBaseUrl)

	srv := service.New(
		storageHdl,
		discoveryHdl,
		srvInfoHdl,
		docClt,
		ladonClt,
		cfg.HttpTimeout,
		cfg.ApiGateway,
		cfg.Filter.AdminRoleName)

	httpHandler, err := api.New(srv, map[string]string{
		api.HeaderApiVer:  srvInfoHdl.GetVersion(),
		api.HeaderSrvName: srvInfoHdl.GetName(),
	})
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	httpServer := util.NewServer(httpHandler, cfg.ServerPort)

	ctx, cf := context.WithCancel(context.Background())

	go func() {
		util.WaitForSignal(ctx, syscall.SIGINT, syscall.SIGTERM)
		cf()
	}()

	if err = storageHdl.Init(ctx); err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.RunPeriodicProcurement(ctx, cfg.Procurement.Interval); err != nil {
			util.Logger.Error(err)
			ec = 1
		}
		cf()
	}()

	go func() {
		if err := util.StartServer(httpServer); err != nil {
			util.Logger.Error(err)
			ec = 1
		}
		cf()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := util.StopServer(ctx, httpServer); err != nil {
			util.Logger.Error(err)
			ec = 1
		}
	}()

	wg.Wait()
}
