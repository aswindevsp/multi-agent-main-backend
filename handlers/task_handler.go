package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"nstorm.com/main-backend/models"
)

type TaskHandler struct {
	db *pgx.Conn
}

func NewTaskHandler(db *pgx.Conn) *TaskHandler {
	return &TaskHandler{db: db}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        INSERT INTO tasks (project_id, assigned_to, title, description, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at`

	err := h.db.QueryRow(ctx, query,
		task.ProjectID,
		task.AssignedTo,
		task.Title,
		task.Description,
		task.Status,
	).Scan(&task.ID, &task.CreatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := `
        SELECT id, project_id, assigned_to, title, description, status, created_at
        FROM tasks 
        WHERE id = $1`

	var task models.Task
	err = h.db.QueryRow(context.Background(), query, taskID).Scan(
		&task.ID,
		&task.ProjectID,
		&task.AssignedTo,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        UPDATE tasks 
        SET project_id = $1, assigned_to = $2, title = $3, description = $4, status = $5
        WHERE id = $6
        RETURNING id, project_id, assigned_to, title, description, status, created_at`

	err = h.db.QueryRow(context.Background(), query,
		task.ProjectID,
		task.AssignedTo,
		task.Title,
		task.Description,
		task.Status,
		taskID,
	).Scan(&task.ID, &task.ProjectID, &task.AssignedTo, &task.Title, &task.Description, &task.Status, &task.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM tasks WHERE id = $1`
	result, err := h.db.Exec(context.Background(), query, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	query := `
        SELECT id, project_id, assigned_to, title, description, status, created_at
        FROM tasks`

	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.AssignedTo,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}