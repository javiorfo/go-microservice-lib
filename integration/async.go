package integration

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/javiorfo/nilo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Async interface {
	Execute(Request)
}

type Try int
type state string

type asyncClient struct {
	try        Try
	collection *mongo.Collection
	client     Client[RawData]
}

func NewAsyncHttpClient(collection *mongo.Collection, try Try) Async {
	return &asyncClient{collection: collection, try: try, client: NewHttpClient[RawData]()}
}

func (a *asyncClient) Execute(r Request) {
	go func() {
		code := uuid.NewString()
		defer func() {
			log.Infof("async %s execution terminated", code)
			if r := recover(); r != nil {
				log.Errorf("async %s. Panic: %v", code, r)
			}
		}()

		model, err := a.create(r.ctx, newAsyncModel(code, r.url, r.body))
		if err != nil {
			log.Errorf("async %s. Error creating mongo model: %v", code, err)
			return
		}
		log.Infof("async %s. Created", code)

		for i := range a.try {
			log.Infof("async %s. Try %d: %s", code, i+1, r.url)
			resp, err := a.client.Send(r)

			if err != nil || resp.Error != nil {
				errStr := nilo.OfPointer(resp).MapToString(func(r Response[RawData]) string {
					return r.ErrorToJson().OrElse("No error response available")
				}).OrElse(err.Error())

				log.Errorf("async %s. Try %d. Error executing request: %v", code, i+1, errStr)

				if err := a.update(r.ctx, model, "ERROR", errStr); err != nil {
					log.Errorf("async %s. Try %d. Error setting error mongo model: %v", code, i+1, err)
					return
				}

				time.Sleep(3 * time.Second)
				continue
			}

			if err = a.update(r.ctx, model, "OK", resp.DataToJson().OrElse("No response available")); err != nil {
				log.Errorf("async %s. Try %d. Error updating mongo model: %v", code, i+1, err)
			} else {
				log.Infof("async %s. Try %d. Succeded", code, i+1)
			}
			return
		}
	}()
}

type asyncModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Code     string             `bson:"code"`
	Endpoint string             `bson:"endpoint"`
	Body     string             `bson:"body"`
	State    state              `bson:"state"`
	Response string             `bson:"response"`
	Date     time.Time          `bson:"date"`
}

func newAsyncModel(code, endpoint string, body *[]byte) asyncModel {
	return asyncModel{
		ID:       primitive.NewObjectID(),
		Code:     code,
		Endpoint: endpoint,
		Body:     string(*body),
		State:    "PROCESSING",
		Date:     time.Now(),
	}
}

func (a asyncClient) create(ctx context.Context, am asyncModel) (asyncModel, error) {
	_, err := a.collection.InsertOne(ctx, am)
	if err != nil {
		return am, err
	}
	return am, nil
}

func (a asyncClient) update(ctx context.Context, am asyncModel, state state, response string) error {
	filter := bson.M{"_id": am.ID}

	am.State = state
	am.Response = response

	update := bson.M{"$set": am}

	_, err := a.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}
