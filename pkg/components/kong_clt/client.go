package kong_clt

import (
	"context"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
	"net/http"
	"net/url"
)

type ClientItf interface {
	GetRoutes(ctx context.Context) ([]Route, error)
	GetServices(ctx context.Context) ([]Service, error)
}

type Client struct {
	baseClient *base_client.Client
	baseUrl    string
	username   string
	password   string
}

func New(httpClient base_client.HTTPClient, baseUrl, username, password string) *Client {
	return &Client{
		baseClient: base_client.New(httpClient, customError, ""),
		baseUrl:    baseUrl,
		username:   username,
		password:   password,
	}
}

func (c *Client) GetRoutes(ctx context.Context) ([]Route, error) {
	u, err := url.JoinPath(c.baseUrl, "routes")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setBasicAuth(req)
	var resp routes
	err = c.baseClient.ExecRequestJSON(req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetServices(ctx context.Context) ([]Service, error) {
	u, err := url.JoinPath(c.baseUrl, "services")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setBasicAuth(req)
	var resp services
	err = c.baseClient.ExecRequestJSON(req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) setBasicAuth(req *http.Request) {
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}
}

func customError(_ int, err error) error {
	return err
}
