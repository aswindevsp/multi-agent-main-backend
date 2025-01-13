package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

type ChatResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Tasks   []TaskAssignment `json:"tasks"`
}

type TaskAssignment struct {
	Task       string `json:"task"`
	AssignedTo string `json:"assigned_to"`
}

func (h *ProjectHandler) GenerateAndAssignTasks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Read the requirements from request body
	var req struct {
		Requirements string `json:"requirements"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Query employees and their skills for the project
	query := `
        SELECT e.name, e.skills
        FROM employees e
        JOIN employee_projects ep ON e.id = ep.employee_id
        WHERE ep.project_id = $1`

	rows, err := h.db.Query(context.Background(), query, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build the prompt with employee skills
	var employeeSkills []string
	employeeNameMap := make(map[string]int)
	for rows.Next() {
		var name string
		var skills []string
		if err := rows.Scan(&name, &skills); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		employeeSkills = append(employeeSkills, fmt.Sprintf("%s: %v", name, skills))
		employeeNameMap[name] = 1
	}

	// Prepare the prompt
	prompt := fmt.Sprintf("Project Requirements: %s\nTeam Members and Skills:\n%s",
		req.Requirements,
		strings.Join(employeeSkills, "\n"))

	// Send request to chat endpoint
	chatReq, err := http.NewRequest("POST", "http://localhost:8080/chat",
		bytes.NewBufferString(fmt.Sprintf(`{"prompt": "%s"}`, prompt)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chatReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var chatResponse ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert tasks into database
	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	insertQuery := `
        INSERT INTO tasks (project_id, title, assigned_to, status)
        SELECT $1, $2, e.id, 'TODO'
        FROM employees e
        WHERE e.name = $3
        RETURNING id`

	for _, task := range chatResponse.Tasks {
		var taskID int
		err := tx.QueryRow(ctx, insertQuery, projectID, task.Task, task.AssignedTo).Scan(&taskID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert task: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResponse)
}
