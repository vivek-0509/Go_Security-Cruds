package handlers

import (
	"awesomeProject/internal/models"
	"awesomeProject/internal/service"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *TaskHandler) TasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListTasks(w, r)
	case http.MethodPost:
		h.CreateTask(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h *TaskHandler) TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetTask(w, r, id)
	case http.MethodPut, http.MethodPatch:
		h.UpdateTask(w, r, id)
	case http.MethodDelete:
		h.DeleteTask(w, r, id)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// Handlers now just call service
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.svc.GetAllTasks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input models.Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input.CreatedAt = time.Now().UTC()
	input.UpdatedAt = time.Now().UTC()

	if err := h.svc.CreateTask(r.Context(), &input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(input)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request, id string) {
	task, err := h.svc.GetTaskByID(r.Context(), id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updated, err := h.svc.UpdateTask(r.Context(), id, updates)
	if err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(updated)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteTask(r.Context(), id); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
