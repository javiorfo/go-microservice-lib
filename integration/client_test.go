package integration

import (
	"context"
	"net/http"
	"testing"
)

type data struct {
	UserId    int    `json:"userId"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func TestIntegrationGet(t *testing.T) {
	client := NewHttpClient[data]()
	resp, err := client.Send(NewRequest(context.Background(), "https://jsonplaceholder.typicode.com/todos/1", WithJsonHeaders()))

	if err != nil {
		t.Error(err.Error())
	}

	if resp.Error != nil {
		t.Error(resp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code must be 200. Got %d", resp.StatusCode)
	}
}

func TestIntegrationInterface(t *testing.T) {
	receiver := func(c Client[RawData]) (*Response[RawData], error) {
		return c.Send(NewRequest(context.Background(), "https://jsonplaceholder.typicode.com/todos/1"))
	}

	resp, err := receiver(NewHttpClient[RawData]())

	if err != nil {
		t.Error(err.Error())
	}

	if resp.Error != nil {
		t.Error(resp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code must be 200. Got %d", resp.StatusCode)
	}
}

func TestIntegrationGetError(t *testing.T) {
	client := NewHttpClient[data]()
	_, err := client.Send(NewRequest(context.Background(), "https://jsonplaceholde/1", WithJsonHeaders()))

	if err == nil {
		t.Fatal("Must be an error")
	}
}

func TestIntegrationPost(t *testing.T) {
	client := NewHttpClient[RawData]()
	resp, err := client.Send(NewRequest(context.Background(), "https://jsonplaceholder.typicode.com/posts",
		WithMethod(http.MethodPost),
		WithHeader("Authorization", "mock_token"),
		WithBody(data{UserId: 100, Title: "test"}),
		WithJsonHeaders(),
	))

	if err != nil {
		t.Error(err.Error())
	}

	data := *resp.Data
	if data["title"] != "test" {
		t.Error("title json property must be 'test'")
	}

	optional := resp.ValueFromJsonField("title")
	if optional.IsNil() {
		t.Error("title json property must be 'test'")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Status code must be 201. Got %d", resp.StatusCode)
	}
}
