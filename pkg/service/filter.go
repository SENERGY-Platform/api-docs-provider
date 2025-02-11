package service

import (
	"context"
	"encoding/json"
	"path"
)

func (s *Service) filterDoc(ctx context.Context, doc map[string]json.RawMessage, userToken string, userRoles []string, basePath string) (bool, error) {
	rawPaths, ok := doc[swaggerPathsKey]
	if !ok {
		return true, nil
	}
	var oldPaths map[string]map[string]json.RawMessage
	err := json.Unmarshal(rawPaths, &oldPaths)
	if err != nil {
		return false, err
	}
	if len(oldPaths) == 0 {
		return true, nil
	}
	var newPaths map[string]map[string]json.RawMessage
	if userToken != "" {

	} else {
		newPaths, err = s.getNewPathsByRoles(ctx, oldPaths, basePath, userRoles)
		if err != nil {
			return false, err
		}
	}
	if len(newPaths) > 0 {
		b, err := json.Marshal(newPaths)
		if err != nil {
			return false, err
		}
		doc[swaggerPathsKey] = b
		return true, nil
	}
	return false, nil
}

func (s *Service) getNewPathsByRoles(ctx context.Context, oldPaths map[string]map[string]json.RawMessage, basePath string, userRoles []string) (map[string]map[string]json.RawMessage, error) {
	newPaths := make(map[string]map[string]json.RawMessage)
	for subPath, methods := range oldPaths {
		allowedMethods, err := s.getAllowedMethodsByRoles(ctx, path.Join(basePath, subPath), userRoles, methods)
		if err != nil {
			return nil, err
		}
		if len(allowedMethods) > 0 {
			newPaths[subPath] = allowedMethods
		}
	}
	return newPaths, nil
}

func (s *Service) getAllowedMethodsByRoles(ctx context.Context, fullPath string, userRoles []string, methods map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	allowedMethods := make(map[string]json.RawMessage)
	for method, rawMessage := range methods {
		for _, role := range userRoles {
			ok, err := s.getAccessPolicyByRole(ctx, fullPath, role, method)
			if err != nil {
				return nil, err
			}
			if ok {
				allowedMethods[method] = rawMessage
				break
			}
		}
	}
	return allowedMethods, nil
}

func (s *Service) getAccessPolicyByRole(ctx context.Context, fullPath, role, method string) (bool, error) {
	ctxWt, cf := context.WithTimeout(ctx, s.timeout)
	defer cf()
	return s.ladonClt.GetRoleAccessPolicy(ctxWt, role, fullPath, method)
}
