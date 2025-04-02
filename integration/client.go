package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type headers map[string]string

type request struct {
	ctx     context.Context
	url     string
	method  string
	headers headers
	body    *[]byte
	err     error
}

type client[T any] struct {
	request    request
	httpClient *http.Client
	response   *Response[T]
}

func NewClient[T any](context context.Context, url string) client[T] {
	return client[T]{
		request: request{
			ctx:     context,
			url:     url,
			headers: make(headers),
		},
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   30 * time.Second,
		},
	}
}

func NewClientRaw(context context.Context, url string) client[RawData] {
	return NewClient[RawData](context, url)
}

func (c client[T]) Method(method string) client[T] {
	c.request.method = method
	return c
}

func (c client[T]) Timeout(timeout time.Duration) client[T] {
	c.httpClient.Timeout = timeout
	return c
}

func (c client[T]) HttpClient(httpClient *http.Client) client[T] {
	c.httpClient = httpClient
	return c
}

func (c client[T]) Header(name, value string) client[T] {
	c.request.headers[name] = value
	return c
}

func (c client[T]) WithCorrelationID() client[T] {
	c.request.headers["x-correlation-id"] = uuid.NewString()
	return c
}

func (c client[T]) WithJsonHeaders() client[T] {
	c.request.headers["Content-Type"] = "application/json"
	c.request.headers["Accept"] = "application/json"
	return c
}

func (c client[T]) Body(entity any) client[T] {
	body, err := json.Marshal(entity)
	if err != nil {
		c.request.err = err
	} else {
		c.request.body = &body
	}
	return c
}

func (c client[T]) Send() (*Response[T], error) {
	req := c.request
	if req.err != nil {
		return nil, req.err
	}

	var bodyBuffer io.Reader
	if req.body != nil {
		bodyBuffer = bytes.NewBuffer(*req.body)
	}

	call, err := http.NewRequestWithContext(req.ctx, req.method, req.url, bodyBuffer)
	if err != nil {
		return nil, err
	}

	for k, v := range req.headers {
		call.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(call)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode < 200 || statusCode > 299 {
		errData, err := decodeData[RawData](resp.Body)
		if err != nil {
			return nil, err
		}

		return &Response[T]{
			StatusCode: statusCode,
			Error:      *errData,
			Headers:    resp.Header,
		}, nil
	}

	data, err := decodeData[T](resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response[T]{
		StatusCode: statusCode,
		Data:       data,
		Headers:    resp.Header,
	}, nil
}

func decodeData[T any](body io.ReadCloser) (*T, error) {
	var data T
	err := json.NewDecoder(body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
