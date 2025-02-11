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
		newPaths, err = s.getNewPathsByToken(ctx, oldPaths, basePath, userToken)
		if err != nil {
			return false, err
		}
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

func (s *Service) getNewPathsByToken(ctx context.Context, oldPaths map[string]map[string]json.RawMessage, basePath string, userToken string) (map[string]map[string]json.RawMessage, error) {
	pathMethodMap := make(map[string][]string)
	for subPath, methods := range oldPaths {
		fullPath := path.Join(basePath, subPath)
		for method := range methods {
			sl := pathMethodMap[fullPath]
			sl = append(sl, method)
			pathMethodMap[fullPath] = sl
		}
	}
	ctxWt, cf := context.WithTimeout(ctx, s.timeout)
	defer cf()
	accessPolicies, err := s.ladonClt.GetUserAccessPolicy(ctxWt, userToken, pathMethodMap)
	if err != nil {
		return nil, err
	}
	newPaths := make(map[string]map[string]json.RawMessage)
	for subPath, methods := range oldPaths {
		allowedMethods := make(map[string]json.RawMessage)
		fullPath := path.Join(basePath, subPath)
		sl, ok := accessPolicies[fullPath]
		if !ok {
			continue
		}
		for _, method := range sl {
			rawMessage, ok := methods[method]
			if !ok {
				continue
			}
			allowedMethods[method] = rawMessage
		}
		if len(allowedMethods) > 0 {
			newPaths[subPath] = allowedMethods
		}
	}
	return newPaths, nil
}

func (s *Service) getNewPathsByRoles(ctx context.Context, oldPaths map[string]map[string]json.RawMessage, basePath string, userRoles []string) (map[string]map[string]json.RawMessage, error) {
	newPaths := make(map[string]map[string]json.RawMessage)
	for subPath, methods := range oldPaths {
		allowedMethods := make(map[string]json.RawMessage)
		fullPath := path.Join(basePath, subPath)
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
		if len(allowedMethods) > 0 {
			newPaths[subPath] = allowedMethods
		}
	}
	return newPaths, nil
}

func (s *Service) getAccessPolicyByRole(ctx context.Context, fullPath, role, method string) (bool, error) {
	ctxWt, cf := context.WithTimeout(ctx, s.timeout)
	defer cf()
	return s.ladonClt.GetRoleAccessPolicy(ctxWt, role, fullPath, method)
}
