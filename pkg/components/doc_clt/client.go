package doc_clt

import (
	"context"
	"errors"
	"fmt"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
	"io"
	"net/http"
	"net/url"
)

type ClientItf interface {
	GetDoc(ctx context.Context, protocol, host string, port int) ([]byte, error)
}

type Client struct {
	baseClient *base_client.Client
	docPath    string
}

func New(httpClient base_client.HTTPClient, docPath string) *Client {
	return &Client{
		baseClient: base_client.New(httpClient, customError, ""),
		docPath:    docPath,
	}
}

func (h *Client) GetDoc(ctx context.Context, protocol, host string, port int) ([]byte, error) {
	baseUrl := fmt.Sprintf("%s://%s", protocol, host)
	if port > 0 {
		baseUrl = baseUrl + fmt.Sprintf(":%d", port)
	}
	u, err := url.JoinPath(baseUrl, h.docPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	body, err := h.baseClient.ExecRequest(req)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("empty response")
	}
	return b, nil
}

func customError(_ int, err error) error {
	return err
}
