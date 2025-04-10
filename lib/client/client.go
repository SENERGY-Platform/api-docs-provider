package client

import (
	"bytes"
	"context"
	"github.com/SENERGY-Platform/api-docs-provider/lib/models"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	baseClient *base_client.Client
	baseUrl    string
}

func New(httpClient base_client.HTTPClient, baseUrl string) *Client {
	return &Client{
		baseClient: base_client.New(httpClient, customError, models.HeaderRequestID),
		baseUrl:    baseUrl,
	}
}

func (c *Client) AsyncapiPutDoc(ctx context.Context, id string, data []byte) error {
	return c.AsyncapiPutDocFromReader(ctx, id, bytes.NewBuffer(data))
}

func (c *Client) AsyncapiPutDocFromReader(ctx context.Context, id string, reader io.Reader) error {
	u, err := url.JoinPath(c.baseUrl, "/storage/asyncapi", id)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, reader)
	if err != nil {
		return err
	}
	return c.baseClient.ExecRequestVoid(req)
}

func customError(code int, err error) error {
	switch code {
	case http.StatusInternalServerError:
		err = models.NewInternalError(err)
	case http.StatusNotFound:
		err = models.NewNotFoundError(err)
	case http.StatusBadRequest:
		err = models.NewInvalidInputError(err)
	case http.StatusConflict:
		err = models.NewResourceBusyError(err)
	}
	return err
}
