package ladon_clt

import (
	"context"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
)

type ClientItf interface {
	GetRoleAccessPolicy(ctx context.Context, role, path, method string) (bool, error)
	GetUserAccessPolicy(ctx context.Context, token string, pathMethodMap map[string][]string) (map[string][]string, error)
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

func customError(_ int, err error) error {
	return err
}
