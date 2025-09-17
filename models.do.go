package main

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrMissingMongoURI = errors.New("missing mongo URI")
)

//task represents a to-do stored in mongoDb

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description,omitempty" json:"description"`
	Completed   bool               `bson:"completed" json:"completed"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
