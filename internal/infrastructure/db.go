package infrastructure

import (
	"context"
	"errors"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Mongo struct {
	Client         *mongo.Client
	Database       *mongo.Database
	TasksColl      *mongo.Collection
	connectionURI  string
	databaseName   string
	collectionName string
}

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

func (m *Mongo) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.connectionURI))
	if err != nil {
		return err
	}

	// Ping to verify
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	m.Client = client
	m.Database = client.Database(m.databaseName)
	m.TasksColl = m.Database.Collection(m.collectionName)
	return nil
}

func (m *Mongo) Close(ctx context.Context) error {
	if m.Client == nil {
		return nil
	}
	return m.Client.Disconnect(ctx)
}
