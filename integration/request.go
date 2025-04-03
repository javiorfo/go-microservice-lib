package integration

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
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

type option struct {
	method  string
	headers headers
	body    *[]byte
	err     error
}

type RequestOptions func(*option)

func WithMethod(method string) RequestOptions {
	return func(o *option) {
		o.method = method
	}
}

func WithHeader(name, value string) RequestOptions {
	return func(o *option) {
		o.headers[name] = value
	}
}

func WithCorrelationID() RequestOptions {
	return func(o *option) {
		o.headers["x-correlation-id"] = uuid.NewString()
	}
}

func WithJsonHeaders() RequestOptions {
	return func(o *option) {
		o.headers["Content-Type"] = "application/json"
		o.headers["Accept"] = "application/json"
	}
}

func WithBody(body any) RequestOptions {
	return func(o *option) {
		body, err := json.Marshal(body)
		if err != nil {
			o.err = err
		} else {
			o.body = &body
		}
	}
}

func NewRequest(context context.Context, url string, options ...RequestOptions) request {
	opts := option{
		headers: make(headers),
	}

	for _, opt := range options {
		opt(&opts)
	}

	return request{
		ctx:     context,
		url:     url,
		headers: opts.headers,
		body:    opts.body,
		err:     opts.err,
		method:  opts.method,
	}
}
