package main

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"os"
	"time"
)

// Holds reference to the mongodb database and collection that we are using
type Mongo struct {
	Client         *mongo.Client
	Database       *mongo.Database
	TasksColl      *mongo.Collection
	connectionURI  string
	databaseName   string
	collectionName string
}

//Created a mongoDb struct with the provided values in env if get connected returns the Mongo Object

func NewMongoFromEnv() (*Mongo, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil, errors.New("MONGO_URI environment variable not set")
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "taskDb"
	}
	collectionName := os.Getenv("MONGO_COLLECTION")
	if collectionName == "" {
		return nil, errors.New("MONGO_COLLECTION environment variable not set")
	}

	m := &Mongo{connectionURI: uri, databaseName: dbName, collectionName: collectionName}
	if err := m.Connect(); err != nil {
		return nil, err
	}
	return m, nil
}

// Method to connect to the db
func (m *Mongo) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, error := mongo.Connect(ctx, options.Client().ApplyURI(m.connectionURI))
	if error != nil {
		return error
	}

	//Ping to verify the connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return err
	}
	m.Client = client
	m.Database = client.Database(m.databaseName)
	m.TasksColl = m.Database.Collection(m.collectionName)
	return nil
}

// method to close the db connection
func (m *Mongo) Close(ctx context.Context) error {
	if m.Client == nil {
		return nil
	}
	return m.Client.Disconnect(ctx)
}
