package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	Client    *mongo.Client
	DB        *mongo.Database
	Forms     *mongo.Collection
	Responses *mongo.Collection
}

func NewMongoStore() (*MongoStore, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "Custom-Form-Builder-with-Live-Analytics"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	store := &MongoStore{
		Client:    client,
		DB:        db,
		Forms:     db.Collection("forms"),
		Responses: db.Collection("responses"),
	}

	_, _ = store.Forms.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetBackground(true),
	})
	_, _ = store.Responses.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "formId", Value: 1}, {Key: "created", Value: -1}},
		Options: options.Index().SetBackground(true),
	})

	log.Printf("connected to MongoDB: %s / db: %s", uri, dbName)
	return store, nil
}
