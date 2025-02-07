package service

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"sync"
)

func (s *Service) refreshDocs(ctx context.Context) error {
	services, err := s.discoveryHdl.GetServices(ctx)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, service := range services {
		if err = ctx.Err(); err != nil {
			break
		}
		if len(service.ExtPaths) > 0 {
			wg.Add(1)
			go s.handleService(ctx, service)
		}
	}
	wg.Wait()
	return nil
}

func (s *Service) handleService(ctx context.Context, service models.Service) {
	ctxWt, cf := context.WithTimeout(ctx, s.timeout)
	defer cf()
	doc, err := s.docClt.GetDoc(ctxWt, service.Protocol, service.Host, service.Port)
	if err != nil {
		return
	}
	if err = s.validateDoc(doc); err != nil {
		util.Logger.Errorf("validating doc for '%s:%d' failed: %s", service.Host, service.Port, err)
		return
	}
	if err = s.storageHdl.Write(ctx, service.ID, service.ExtPaths, doc); err != nil {
		util.Logger.Errorf("writing doc for '%s:%d' failed: %s", service.Host, service.Port, err)
		return
	}
}

func (s *Service) validateDoc(doc []byte) error {
	return json.Unmarshal(doc, &map[string]json.RawMessage{})
}
