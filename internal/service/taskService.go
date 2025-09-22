package service

import (
	"awesomeProject/internal/models"
	"awesomeProject/internal/repository"
	"context"
)

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	return s.repo.FindAll(ctx)
}

func (s *TaskService) GetTaskByID(ctx context.Context, id string) (*models.Task, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *TaskService) CreateTask(ctx context.Context, task *models.Task) error {
	return s.repo.Create(ctx, task)
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, updates map[string]interface{}) (*models.Task, error) {
	return s.repo.Update(ctx, id, updates)
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
