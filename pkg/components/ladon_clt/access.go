package ladon_clt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

func (c *Client) GetRoleAccessPolicy(ctx context.Context, role, path, method string) (bool, error) {
	u, err := url.JoinPath(c.baseUrl, "access")
	if err != nil {
		return false, err
	}
	body, err := json.Marshal(roleAccessRequest{
		Resource: "endpoints" + strings.ReplaceAll(path, "/", ":"),
		Action:   strings.ToUpper(method),
		Subject:  role,
	})
	if err != nil {
		return false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	var res roleAccessResponse
	err = c.baseClient.ExecRequestJSON(req, &res)
	if err != nil {
		return false, err
	}
	return res.Result, nil
}

func (c *Client) GetUserAccessPolicy(ctx context.Context, token string, pathMethodMap map[string][]string) (map[string][]string, error) {
	u, err := url.JoinPath(c.baseUrl, "allowed")
	if err != nil {
		return nil, err
	}
	aReq := newUserAccessRequest(pathMethodMap)
	body, err := json.Marshal(aReq)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	var res userAccessResponse
	err = c.baseClient.ExecRequestJSON(req, &res)
	if err != nil {
		return nil, err
	}
	return getUserAccessResult(aReq, res.Allowed)
}

func getUserAccessResult(request []userAccessRequest, response []bool) (map[string][]string, error) {
	if len(request) != len(response) {
		return nil, errors.New("bad response")
	}
	result := make(map[string][]string)
	for i, item := range request {
		sl := result[item.orgEndpoint]
		if ok := response[i]; ok {
			sl = append(sl, item.orgMethod)
		}
		result[item.orgEndpoint] = sl
	}
	return result, nil
}

func newUserAccessRequest(pathMethodMap map[string][]string) []userAccessRequest {
	var request []userAccessRequest
	for p, methods := range pathMethodMap {
		for _, method := range methods {
			ar := userAccessRequest{
				Method:      strings.ToUpper(method),
				orgMethod:   method,
				orgEndpoint: p,
			}
			if !strings.HasPrefix(p, "/") {
				ar.Endpoint = "/" + p
			} else {
				ar.Endpoint = p
			}
			request = append(request, ar)
		}
	}
	slices.SortStableFunc(request, func(a, b userAccessRequest) int {
		return strings.Compare(a.Endpoint+a.Method, b.Endpoint+b.Method)
	})
	return request
}
