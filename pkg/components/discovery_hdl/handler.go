package discovery_hdl

import (
	"context"
	"fmt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/kong_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"time"
)

type Handler struct {
	kongClient    kong_clt.ClientItf
	timeout       time.Duration
	hostBlacklist map[string]struct{}
}

func New(kongClient kong_clt.ClientItf, timeout time.Duration, hostBlacklist []string) *Handler {
	blackList := make(map[string]struct{})
	for _, host := range hostBlacklist {
		blackList[host] = struct{}{}
	}
	return &Handler{
		kongClient:    kongClient,
		timeout:       timeout,
		hostBlacklist: blackList,
	}
}

func (h *Handler) GetServices(ctx context.Context) (map[string]models.Service, error) {
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	kRoutes, err := h.kongClient.GetRoutes(ctxWt)
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	ctxWt2, cf2 := context.WithTimeout(ctx, h.timeout)
	defer cf2()
	kServices, err := h.kongClient.GetServices(ctxWt2)
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	kSrvMap := getKongSrvMap(kServices)
	services := make(map[string]models.Service)
	for _, kRoute := range kRoutes {
		if len(kRoute.Paths) == 0 {
			continue
		}
		kService, ok := kSrvMap[kRoute.Service.ID]
		if !ok {
			continue
		}
		if _, ok = h.hostBlacklist[kService.Host]; ok {
			continue
		}
		id := fmt.Sprintf("%s%d", kService.Host, kService.Port)
		service, ok := services[id]
		if !ok {
			service.ID = id
			service.Host = kService.Host
			service.Port = kService.Port
			service.Protocol = kService.Protocol
		}
		service.ExtPaths = append(service.ExtPaths, kRoute.Paths...)
		services[id] = service
	}
	for _, service := range services {
		util.Logger.Debugf("discovery: found service host='%s' port=%d for %v", service.Host, service.Port, service.ExtPaths)
	}
	return services, nil
}

func getKongSrvMap(kServices []kong_clt.Service) map[string]kong_clt.Service {
	srvMap := make(map[string]kong_clt.Service)
	for _, kService := range kServices {
		srvMap[kService.ID] = kService
	}
	return srvMap
}
