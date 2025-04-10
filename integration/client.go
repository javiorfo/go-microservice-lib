package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client[T any] interface {
	Send(Request) (*Response[T], error)
}

type client[T any] struct {
	client *http.Client
}

func NewHttpClient[T any]() client[T] {
	return NewHttpClientWithTimeout[T](30)
}

func NewHttpClientWithTimeout[T any](timeout time.Duration) client[T] {
	return client[T]{
		client: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   timeout * time.Second,
		},
	}
}

func (c client[T]) Send(req Request) (*Response[T], error) {
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

	resp, err := c.client.Do(call)
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
