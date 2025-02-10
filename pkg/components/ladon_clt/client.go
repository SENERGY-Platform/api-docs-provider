package ladon_clt

import (
	"bytes"
	"context"
	"encoding/json"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
	"net/http"
	"net/url"
	"strings"
)

type ClientItf interface {
	GetAccessPolicy(ctx context.Context, role, path, method string) (bool, error)
}

type Client struct {
	baseClient *base_client.Client
	baseUrl    string
}

func New(httpClient base_client.HTTPClient, baseUrl string) *Client {
	return &Client{
		baseClient: base_client.New(httpClient, customError, ""),
		baseUrl:    baseUrl,
	}
}

func (c *Client) GetAccessPolicy(ctx context.Context, role, path, method string) (bool, error) {
	u, err := url.JoinPath(c.baseUrl, "access")
	if err != nil {
		return false, err
	}
	body, err := json.Marshal(accessRequest{
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
	var res accessResponse
	err = c.baseClient.ExecRequestJSON(req, &res)
	if err != nil {
		return false, err
	}
	return res.Result, nil
}

func customError(_ int, err error) error {
	return err
}
