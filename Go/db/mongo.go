package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Client  *mongo.Client
	Context context.Context
}

func EstablishConnection(ctx context.Context) (*Database, error) {
	uri, err := getMongoURI()
	if err != nil {
		return nil, err
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if _, err = client.Database("blog").Collection("posts").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{"year", 1}, {"month", 1}, {"day", 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return nil, err
	}
	return &Database{
		Client:  client,
		Context: ctx,
	}, nil
}

func getMongoURI() (string, error) {
	username, ok := os.LookupEnv("DB_USERNAME")
	if !ok {
		return "", errors.New("DB_USERNAME env variable missing")
	}
	password, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		return "", errors.New("DB_PASSWORD env variable missing")
	}
	return fmt.Sprintf("mongodb://%s:%s@host.docker.internal:27017", username, password), nil
}
