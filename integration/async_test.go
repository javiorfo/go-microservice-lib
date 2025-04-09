package integration

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var async Async

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp"),
	}

	mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start container: %s", err)
	}
	defer mongoContainer.Terminate(ctx)

	host, err := mongoContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %s", err)
	}
	port, err := mongoContainer.MappedPort(ctx, "27017")
	if err != nil {
		log.Fatalf("Failed to get container port: %s", err)
	}

	clientOptions := options.Client().ApplyURI("mongodb://" + host + ":" + port.Port())
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %s", err)
	}
	collection := client.Database("testdb").Collection("dummies")

	async = NewAsyncHttpClient(collection, 3)

	code := m.Run()

	os.Exit(code)
}

func TestAsync(t *testing.T) {
	ctx := context.Background()
	request := NewRequest(ctx, "https://jsonplaceholder.typicode.com/posts",
		WithMethod(http.MethodPost),
		WithBody(data{UserId: 100, Title: "test"}),
		WithJsonHeaders(),
	)

	start := time.Now()
	async.Execute(request)
	elapsed := time.Since(start)
	t.Logf("Execution took %s\n", elapsed)
}

func TestAsyncError(t *testing.T) {
	ctx := context.Background()
	request := NewRequest(ctx, "https://jsonplacehol",
		WithMethod(http.MethodPost),
		WithBody(data{UserId: 100, Title: "test"}),
		WithJsonHeaders(),
	)

	start := time.Now()
	async.Execute(request)
	elapsed := time.Since(start)
	t.Logf("Execution took %s\n", elapsed)

	// Wait 20 secs to check the retries
	time.Sleep(20 * time.Second)
}
