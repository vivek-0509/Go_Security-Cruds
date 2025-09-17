package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	db *Mongo
}

func NewServer(db *Mongo) *Server {
	return &Server{db: db}
}

// Routes registers HTTP handlers using net/http no framework used
func (s *Server) Routes() {
	http.HandleFunc("/healthz", s.healthHandler)
	http.HandleFunc("/tasks", s.tasksHandler)
	http.HandleFunc("/tasks/", s.taskByIdHandler)
}

// just a health check function for client to know that server is properly running and responding , it is just a good practise to use not a necessary thing to do
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).
		Encode(map[string]string{"status": "ok"})
}

// tasksHandler handles collection routes: GET /tasks, POST /tasks
func (s *Server) tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTasks(w, r)
	case http.MethodPost:
		s.createTask(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// taskByIdHandler handles item routes: GET/PATCH/PUT/POST
func (s *Server) taskByIdHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getTask(w, r, id)
	case http.MethodPut, http.MethodPatch:
		s.updateTask(w, r, id)
	case http.MethodDelete:
		s.deleteTask(w, r, id)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	curr, err := s.db.TasksColl.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(bson.M{"createdAt": -1}),
	)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	defer curr.Close(ctx)

	var tasks []Task
	for curr.Next(ctx) {
		var t Task
		if err := curr.Decode(&t); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}

	if err := curr.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tasks)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Completed   *bool  `json:"completed"`
	}
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()

	task := Task{
		ID:          primitive.NewObjectID(),
		Title:       input.Title,
		Description: input.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if input.Completed != nil {
		task.Completed = *input.Completed
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	_, err = s.db.TasksColl.InsertOne(ctx, task)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(task)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request, id string) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var task Task
	err = s.db.TasksColl.FindOne(ctx, bson.M{"_id": oid}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request, id string) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var input struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Completed   *bool   `json:"completed"`
	}
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	update := bson.M{}
	if input.Title != nil {
		if strings.TrimSpace(*input.Title) == "" {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		update["title"] = *input.Title
	}
	if input.Description != nil {
		update["description"] = *input.Description
	}
	if input.Completed != nil {
		update["completed"] = *input.Completed
	}
	if len(update) == 0 {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	update["updatedAt"] = time.Now().UTC()
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated Task
	err = s.db.TasksColl.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": update}, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request, id string) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	res, err := s.db.TasksColl.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.DeletedCount == 0 {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

