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
	resp, err := NewClient[data](context.Background(), "https://jsonplaceholder.typicode.com/todos/1").
		WithJsonHeaders().Send()

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
	_, err := NewClient[data](context.Background(), "https://jsonplaceholde/1").
		WithJsonHeaders().Send()

	if err == nil {
		t.Fatal("Must be an error")
	}
}

func TestIntegrationPost(t *testing.T) {
	resp, err := NewClientRaw(context.Background(), "https://jsonplaceholder.typicode.com/posts").
		Method(http.MethodPost).
		Header("Authorization", "mock_token").
		Body(data{UserId: 100, Title: "test"}).
		WithJsonHeaders().
		Send()

	if err != nil {
		t.Error(err.Error())
	}

	data := *resp.Data
	if data["title"] != "test" {
		t.Error("title json property must be 'test'")
	}

	optional := resp.ValueFromJsonField("title")
	if optional.IsEmpty() {
		t.Error("title json property must be 'test'")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Status code must be 201. Got %d", resp.StatusCode)
	}
}
