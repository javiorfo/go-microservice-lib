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

func (r Response[T]) ValueFromJsonField(jsonField string) nilo.Optional[any] {
	if r.Data == nil {
		return nilo.Empty[any]()
	}

	mapper, isRawData := any(*r.Data).(RawData)
	if !isRawData {
		return nilo.Empty[any]()
	}

	result, ok := mapper[jsonField]
	if !ok {
		return nilo.Empty[any]()
	}

	return nilo.Of(result)
}

func (r Response[T]) ErrorToJson() nilo.Optional[string] {
	if r.Error == nil {
		return nilo.Empty[string]()
	}
	jsonBytes, err := json.Marshal(r.Error)
	if err != nil {
		return nilo.Empty[string]()
	}
	return nilo.Of(string(jsonBytes))
}

func (r Response[T]) DataToJson() nilo.Optional[string] {
	jsonBytes, err := json.Marshal(*r.Data)
	if err != nil {
		return nilo.Empty[string]()
	}
	return nilo.Of(string(jsonBytes))
}
