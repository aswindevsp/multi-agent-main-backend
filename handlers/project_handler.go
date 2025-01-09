package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"nstorm.com/main-backend/models"
)

type ProjectHandler struct {
	db *pgx.Conn
}

func NewProjectHandler(db *pgx.Conn) *ProjectHandler {
	return &ProjectHandler{db: db}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var project models.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(project.LeadID)
	query := `
        INSERT INTO projects (name, description, lead_id)
        VALUES ($1, $2, $3)
        RETURNING id, created_at`

	err := h.db.QueryRow(ctx, query,
		project.Name,
		project.Description,
		project.LeadID,
	).Scan(&project.ID, &project.CreatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// GetProjectByID retrieves a project by its ID
func (h *ProjectHandler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	query := `
        SELECT id, name, description, lead_id, created_at 
        FROM projects 
        WHERE id = $1`

	var project models.Project
	err = h.db.QueryRow(context.Background(), query, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.LeadID,
		&project.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// UpdateProject updates an existing project
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var project models.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        UPDATE projects 
        SET name = $1, description = $2, lead_id = $3
        WHERE id = $4
        RETURNING id, name, description, lead_id, created_at`

	err = h.db.QueryRow(context.Background(), query,
		project.Name,
		project.Description,
		project.LeadID,
		projectID,
	).Scan(&project.ID, &project.Name, &project.Description, &project.LeadID, &project.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// DeleteProject deletes a project by its ID
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM projects WHERE id = $1`
	result, err := h.db.Exec(context.Background(), query, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetAllProjects retrieves all projects
func (h *ProjectHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	query := `
        SELECT id, name, description, lead_id, created_at 
        FROM projects`

	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.LeadID,
			&project.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}
