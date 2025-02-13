package service

import (
	"context"
	"encoding/json"
	"path"
	"regexp"
)

var regRegex = regexp.MustCompile(`\"\$ref\": ?\"#\/definitions\/(.+)\"`)

func (s *Service) filterDoc(ctx context.Context, doc map[string]json.RawMessage, userToken string, userRoles []string, basePath string) (bool, error) {
	oldPaths, err := getDocPaths(doc)
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
	if len(newPaths) == 0 {
		return false, nil
	}
	if err = setDocPaths(doc, newPaths); err != nil {
		return false, err
	}
	oldDefs, err := getDocDefs(doc)
	if err != nil {
		return false, err
	}
	if len(oldDefs) == 0 {
		return true, nil
	}
	if err = setDocDefs(doc, getNewDefinitions(oldDefs, allowedRefs)); err != nil {
		return false, err
	}
	return true, nil
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

func getDocPaths(doc map[string]json.RawMessage) (map[string]map[string]json.RawMessage, error) {
	rawPaths, ok := doc[swaggerPathsKey]
	if !ok {
		return nil, nil
	}
	var paths map[string]map[string]json.RawMessage
	err := json.Unmarshal(rawPaths, &paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func setDocPaths(doc map[string]json.RawMessage, newPaths map[string]map[string]json.RawMessage) error {
	b, err := json.Marshal(newPaths)
	if err != nil {
		return err
	}
	doc[swaggerPathsKey] = b
	return nil
}

func getDocDefs(doc map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	rawDefs, ok := doc[swaggerDefinitionsKey]
	if !ok {
		return nil, nil
	}
	var defs map[string]json.RawMessage
	if err := json.Unmarshal(rawDefs, &defs); err != nil {
		return nil, err
	}
	return defs, nil
}

func setDocDefs(doc map[string]json.RawMessage, newDefs map[string]json.RawMessage) error {
	b, err := json.Marshal(newDefs)
	if err != nil {
		return err
	}
	doc[swaggerDefinitionsKey] = b
	return nil
}

func getNewDefinitions(oldDefs map[string]json.RawMessage, allowedRefs map[string]struct{}) map[string]json.RawMessage {
	for ref, rawMessage := range oldDefs {
		if _, ok := allowedRefs[ref]; ok {
			getDefinitionRefs(rawMessage, allowedRefs)
		}
	}
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
