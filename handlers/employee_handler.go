package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
	"nstorm.com/main-backend/models"
)

type EmployeeHandler struct {
	db *pgx.Conn
}

func NewEmployeeHandler(db *pgx.Conn) *EmployeeHandler {
	return &EmployeeHandler{db: db}
}

func (h *EmployeeHandler) GetEmployeeTasks(w http.ResponseWriter, r *http.Request) {
	query := `
        SELECT 
            e.id as employee_id,
            e.name as employee_name,
            t.id as task_id,
            t.title as task_title,
            t.status as task_status
        FROM employees e
        LEFT JOIN tasks t ON e.id = t.assigned_to
        ORDER BY e.id, t.id`

	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	employeeMap := make(map[int]*models.Employee)

	for rows.Next() {
		var empID int
		var empName string
		var taskID *int
		var taskTitle, taskStatus *string

		err := rows.Scan(&empID, &empName, &taskID, &taskTitle, &taskStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, exists := employeeMap[empID]; !exists {
			employeeMap[empID] = &models.Employee{
				ID:    empID,
				Name:  empName,
				Tasks: []models.Task{},
			}
		}

		if taskID != nil {
			task := models.Task{
				ID:     *taskID,
				Title:  *taskTitle,
				Status: *taskStatus,
			}
			employeeMap[empID].Tasks = append(employeeMap[empID].Tasks, task)
		}
	}

	employees := make([]*models.Employee, 0, len(employeeMap))
	for _, emp := range employeeMap {
		employees = append(employees, emp)
	}

	json.NewEncoder(w).Encode(employees)
}

func (h *EmployeeHandler) AssignToProject(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EmployeeID int `json:"employee_id"`
		ProjectID  int `json:"project_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        INSERT INTO employee_projects (employee_id, project_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING`

	_, err := h.db.Exec(context.Background(), query, request.EmployeeID, request.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EmployeeHandler) RemoveFromProject(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EmployeeID int `json:"employee_id"`
		ProjectID  int `json:"project_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        DELETE FROM employee_projects 
        WHERE employee_id = $1 AND project_id = $2`

	result, err := h.db.Exec(context.Background(), query, request.EmployeeID, request.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Employee not assigned to project", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EmployeeHandler) AssignTask(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TaskID     int `json:"task_id"`
		EmployeeID int `json:"employee_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        UPDATE tasks 
        SET assigned_to = $1 
        WHERE id = $2
        RETURNING id`

	var taskID int
	err := h.db.QueryRow(context.Background(), query, request.EmployeeID, request.TaskID).Scan(&taskID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EmployeeHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TaskID int `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        UPDATE tasks 
        SET status = 'COMPLETED' 
        WHERE id = $1
        RETURNING id`

	var taskID int
	err := h.db.QueryRow(context.Background(), query, request.TaskID).Scan(&taskID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
