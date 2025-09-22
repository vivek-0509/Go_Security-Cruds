package infrastructure

import (
	"awesomeProject/internal/models"
	"awesomeProject/internal/repository"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoTaskRepository struct {
	coll *mongo.Collection
}

func NewMongoTaskRepository(db *Mongo) repository.TaskRepository {
	return &MongoTaskRepository{coll: db.TasksColl}
}

func (m *MongoTaskRepository) FindAll(ctx context.Context) ([]models.Task, error) {
	cur, err := m.coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var tasks []models.Task
	for cur.Next(ctx) {
		var t models.Task
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, cur.Err()
}

func (m *MongoTaskRepository) FindByID(ctx context.Context, id string) (*models.Task, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var task models.Task
	if err := m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (m *MongoTaskRepository) Create(ctx context.Context, task *models.Task) error {
	task.ID = primitive.NewObjectID()
	task.CreatedAt = time.Now().UTC()
	task.UpdatedAt = time.Now().UTC()
	_, err := m.coll.InsertOne(ctx, task)
	return err
}

func (m *MongoTaskRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*models.Task, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	updates["updatedAt"] = time.Now().UTC()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.Task
	if err := m.coll.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": updates}, opts).Decode(&updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (m *MongoTaskRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
