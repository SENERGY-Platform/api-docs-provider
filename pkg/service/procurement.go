package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"sync"
	"time"
)

func (s *Service) RunPeriodicProcurement(ctx context.Context, interval time.Duration) error {
	util.Logger.Info("starting periodic procurement")
	var lErr error
	defer func() {
		if r := recover(); r != nil {
			lErr = fmt.Errorf("periodic procurement paniced:\n%v", r)
		}
		util.Logger.Info("periodic procurement halted")
	}()
	timer := time.NewTimer(time.Microsecond)
	loop := true
	for loop {
		select {
		case <-timer.C:
			err := s.refreshDocs(ctx)
			if err != nil {
				util.Logger.Errorf("procurement failed: %s", err)
			}
			timer.Reset(interval)
		case <-ctx.Done():
			loop = false
			util.Logger.Info("stopping periodic procurement")
			break
		}
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	return lErr
}

func (s *Service) refreshDocs(ctx context.Context) error {
	services, err := s.discoveryHdl.GetServices(ctx)
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}
	for _, service := range services {
		if err = ctx.Err(); err != nil {
			break
		}
		if len(service.ExtPaths) > 0 {
			wg.Add(1)
			go s.handleService(ctx, wg, service)
		}
	}
	wg.Wait()
	reqID := util.GetReqID(ctx)
	if err = s.cleanOldServices(ctx, services); err != nil {
		util.Logger.Errorf("serivce: %sremoving old docs failed: %s", reqID, err)
	}
	return nil
}

func (s *Service) cleanOldServices(ctx context.Context, services map[string]models.Service) error {
	storedServices, err := s.storageHdl.List(ctx)
	if err != nil {
		return err
	}
	reqID := util.GetReqID(ctx)
	for _, service := range storedServices {
		if _, ok := services[service.ID]; !ok {
			if err = s.storageHdl.Delete(ctx, service.ID); err != nil {
				util.Logger.Errorf("serivce: %sremoving old doc failed: %s", reqID, err)
			}
		}
	}
	return nil
}

func (s *Service) handleService(ctx context.Context, wg *sync.WaitGroup, service models.Service) {
	defer wg.Done()
	ctxWt, cf := context.WithTimeout(ctx, s.timeout)
	defer cf()
	doc, err := s.docClt.GetDoc(ctxWt, service.Protocol, service.Host, service.Port)
	if err != nil {
		return
	}
	reqID := util.GetReqID(ctx)
	if err = s.validateDoc(doc); err != nil {
		util.Logger.Warningf("service: %svalidating doc for '%s:%d' failed: %s", reqID, service.Host, service.Port, err)
		return
	}
	if err = s.storageHdl.Write(ctx, service.ID, service.ExtPaths, doc); err != nil {
		util.Logger.Errorf("service: %swriting doc for '%s:%d' failed: %s", reqID, service.Host, service.Port, err)
		return
	}
}

func (s *Service) validateDoc(doc []byte) error {
	var tmp map[string]json.RawMessage
	if err := json.Unmarshal(doc, &tmp); err != nil {
		return err
	}
	for key := range tmp {
		if _, ok := commonSwaggerKeys[key]; ok {
			return nil
		}
	}
	return fmt.Errorf("missing common swagger keys")
}
