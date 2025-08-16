package integration

import (
	"encoding/json"
	"net/http"

	"github.com/javiorfo/nilo"
)

type RawData map[string]any

type Response[T any] struct {
	StatusCode int
	Data       *T
	Error      RawData
	Headers    http.Header
}

func (r Response[T]) ValueFromJsonField(jsonField string) nilo.Option[any] {
	if r.Data == nil {
		return nilo.None[any]()
	}

	mapper, isRawData := any(*r.Data).(RawData)
	if !isRawData {
		return nilo.None[any]()
	}

	result, ok := mapper[jsonField]
	if !ok {
		return nilo.None[any]()
	}

	return nilo.Some(result)
}

func (r Response[T]) ErrorToJson() nilo.Option[string] {
	if r.Error == nil {
		return nilo.None[string]()
	}
	jsonBytes, err := json.Marshal(r.Error)
	if err != nil {
		return nilo.None[string]()
	}
	return nilo.Some(string(jsonBytes))
}

func (r Response[T]) DataToJson() nilo.Option[string] {
	jsonBytes, err := json.Marshal(*r.Data)
	if err != nil {
		return nilo.None[string]()
	}
	return nilo.Some(string(jsonBytes))
}
