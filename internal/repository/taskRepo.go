package repository

import (
	"awesomeProject/internal/models"
	"context"
)

type TaskRepository interface {
	FindAll(ctx context.Context) ([]models.Task, error)
	FindByID(ctx context.Context, id string) (*models.Task, error)
	Create(ctx context.Context, task *models.Task) error
	Update(ctx context.Context, id string, updates map[string]interface{}) (*models.Task, error)
	Delete(ctx context.Context, id string) error
}
