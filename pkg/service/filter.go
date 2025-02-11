package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
)

var regRegex = regexp.MustCompile(`\"\$ref\": ?\"#\/definitions\/(.+)\"`)

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
	var allowedRefs map[string]struct{}
	if userToken != "" {
		newPaths, allowedRefs, err = s.getNewPathsByToken(ctx, oldPaths, basePath, userToken)
		if err != nil {
			return false, err
		}
	} else {
		newPaths, allowedRefs, err = s.getNewPathsByRoles(ctx, oldPaths, basePath, userRoles)
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
	rawDefs, ok := doc[swaggerDefinitionsKey]
	if !ok {
		return true, nil
	}
	var oldDefs map[string]json.RawMessage
	if err = json.Unmarshal(rawDefs, &oldDefs); err != nil {
		return false, err
	}
	newDefs := getNewDefinitions(oldDefs, allowedRefs)
	b, err := json.Marshal(newDefs)
	if err != nil {
		return false, err
	}
	doc[swaggerDefinitionsKey] = b
	return false, nil
}

func (s *Service) getNewPathsByToken(ctx context.Context, oldPaths map[string]map[string]json.RawMessage, basePath string, userToken string) (map[string]map[string]json.RawMessage, map[string]struct{}, error) {
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
		return nil, nil, err
	}
	newPaths := make(map[string]map[string]json.RawMessage)
	defRefs := make(map[string]struct{})
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
			getDefinitionRefs(rawMessage, defRefs)
		}
		if len(allowedMethods) > 0 {
			newPaths[subPath] = allowedMethods
		}
	}
	return newPaths, defRefs, nil
}

func (s *Service) getNewPathsByRoles(ctx context.Context, oldPaths map[string]map[string]json.RawMessage, basePath string, userRoles []string) (map[string]map[string]json.RawMessage, map[string]struct{}, error) {
	newPaths := make(map[string]map[string]json.RawMessage)
	defRefs := make(map[string]struct{})
	for subPath, methods := range oldPaths {
		allowedMethods := make(map[string]json.RawMessage)
		fullPath := path.Join(basePath, subPath)
		for method, rawMessage := range methods {
			for _, role := range userRoles {
				ok, err := s.getAccessPolicyByRole(ctx, fullPath, role, method)
				if err != nil {
					return nil, nil, err
				}
				if ok {
					allowedMethods[method] = rawMessage
					getDefinitionRefs(rawMessage, defRefs)
					break
				}
			}
		}
		if len(allowedMethods) > 0 {
			newPaths[subPath] = allowedMethods
		}
	}
	return newPaths, defRefs, nil
}

func (s *Service) getAccessPolicyByRole(ctx context.Context, fullPath, role, method string) (bool, error) {
	ctxWt, cf := context.WithTimeout(ctx, s.timeout)
	defer cf()
	return s.ladonClt.GetRoleAccessPolicy(ctxWt, role, fullPath, method)
}

func getNewDefinitions(oldDefs map[string]json.RawMessage, allowedRefs map[string]struct{}) map[string]json.RawMessage {
	fmt.Println(allowedRefs)
	for ref, rawMessage := range oldDefs {
		if _, ok := allowedRefs[ref]; ok {
			getDefinitionRefs(rawMessage, allowedRefs)
		}
	}
	fmt.Println(allowedRefs)
	newDefs := make(map[string]json.RawMessage)
	for ref := range allowedRefs {
		if rawMessage, ok := oldDefs[ref]; ok {
			newDefs[ref] = rawMessage
		}
	}
	return newDefs
}

func getDefinitionRefs(raw []byte, refs map[string]struct{}) {
	res := regRegex.FindAllSubmatch(raw, -1)
	for _, re := range res {
		if len(re) > 1 {
			refs[string(re[1])] = struct{}{}
		}
	}
}
