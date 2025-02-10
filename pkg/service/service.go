package service

import (
	"context"
	"encoding/json"
	srv_info_hdl "github.com/SENERGY-Platform/go-service-base/srv-info-hdl"
	srv_info_lib "github.com/SENERGY-Platform/go-service-base/srv-info-hdl/lib"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/doc_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/components/ladon_clt"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util"
	"slices"
	"strings"
	"sync"
	"time"
)

type Service struct {
	storageHdl    StorageHandler
	discoveryHdl  DiscoveryHandler
	srvInfoHdl    srv_info_hdl.SrvInfoHandler
	docClt        doc_clt.ClientItf
	ladonClt      ladon_clt.ClientItf
	timeout       time.Duration
	apiGtwHost    string
	adminRoleName string
}

func New(storageHdl StorageHandler, discoveryHdl DiscoveryHandler, srvInfoHdl srv_info_hdl.SrvInfoHandler, docClt doc_clt.ClientItf, ladonClt ladon_clt.ClientItf, timeout time.Duration, apiGtwHost string, adminRoleName string) *Service {
	return &Service{
		storageHdl:    storageHdl,
		discoveryHdl:  discoveryHdl,
		srvInfoHdl:    srvInfoHdl,
		docClt:        docClt,
		ladonClt:      ladonClt,
		timeout:       timeout,
		apiGtwHost:    apiGtwHost,
		adminRoleName: adminRoleName,
	}
}

func (s *Service) GetSwaggerDocs(ctx context.Context, userRoles []string) ([]map[string]json.RawMessage, error) {
	if len(userRoles) == 0 {
		return []map[string]json.RawMessage{}, nil
	}
	data, err := s.storageHdl.List(ctx)
	if err != nil {
		return []map[string]json.RawMessage{}, models.NewInternalError(err)
	}
	isAdmin := stringInSlice(s.adminRoleName, userRoles)
	var docWrappers []docWrapper
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for _, item := range data {
		rawDoc, err := s.storageHdl.Read(ctx, item.ID)
		if err != nil {
			util.Logger.Errorf("reading swagger doc for %v failed: %s", item.ExtPaths, err)
			continue
		}
		wg.Add(1)
		go func(rawDoc []byte, extPaths []string) {
			defer wg.Done()
			for _, basePath := range extPaths {
				util.Logger.Debugf("transforming swagger doc for '%s'", basePath)
				doc, err := s.transformDoc(rawDoc, basePath)
				if err != nil {
					util.Logger.Errorf("transforming swagger doc for '%s' failed: %s", basePath, err)
					continue
				}
				if !isAdmin {
					ok, err := s.filterDoc(ctx, doc, userRoles, basePath)
					if err != nil {
						util.Logger.Errorf("filtering swagger doc for '%s' failed: %s", basePath, err)
						continue
					}
					if !ok {
						continue
					}
				}
				mu.Lock()
				docWrappers = append(docWrappers, docWrapper{basePath: basePath, doc: doc})
				mu.Unlock()
				util.Logger.Debugf("appended swagger doc for '%s'", basePath)
			}
		}(rawDoc, item.ExtPaths)
	}
	wg.Wait()
	slices.SortStableFunc(docWrappers, func(a, b docWrapper) int {
		return strings.Compare(a.basePath, b.basePath)
	})
	docs := make([]map[string]json.RawMessage, 0, len(docWrappers))
	for _, dw := range docWrappers {
		docs = append(docs, dw.doc)
	}
	return docs, nil
}

func (s *Service) HealthCheck(ctx context.Context) error {
	if _, err := s.storageHdl.List(ctx); err != nil {
		return models.NewInternalError(err)
	}
	return nil
}

func (s *Service) GetSrvInfo(_ context.Context) srv_info_lib.SrvInfo {
	return s.srvInfoHdl.GetInfo()
}

func (s *Service) transformDoc(rawDoc []byte, basePath string) (map[string]json.RawMessage, error) {
	var doc map[string]json.RawMessage
	err := json.Unmarshal(rawDoc, &doc)
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	b, err := json.Marshal(s.apiGtwHost)
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	doc[swaggerHostKey] = b
	b, err = json.Marshal(basePath)
	if err != nil {
		return nil, models.NewInternalError(err)
	}
	doc[swaggerBasePathKey] = b
	if _, ok := doc[swaggerSchemesKey]; !ok {
		b, err = json.Marshal([]string{"https"})
		if err != nil {
			return nil, models.NewInternalError(err)
		}
		doc[swaggerSchemesKey] = b
	}
	return doc, nil
}

func stringInSlice(a string, sl []string) bool {
	for _, b := range sl {
		if b == a {
			return true
		}
	}
	return false
}
