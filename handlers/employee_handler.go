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

type EmployeeHandler struct {
	db *pgx.Conn
}

func NewEmployeeHandler(db *pgx.Conn) *EmployeeHandler {
	return &EmployeeHandler{db: db}
}

func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var employee models.Employee
	if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        INSERT INTO employees (name, email, role, skills)
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, email, role, skills, created_at`

	err := h.db.QueryRow(
		context.Background(),
		query,
		employee.Name,
		employee.Email,
		employee.Role,
		employee.Skills,
	).Scan(
		&employee.ID,
		&employee.Name,
		&employee.Email,
		&employee.Role,
		&employee.Skills,
		&employee.CreatedAt,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(employee)
}

func (h *EmployeeHandler) GetEmployeeById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var employee models.Employee
	query := `
        SELECT id, name, email, role, skills, created_at
        FROM employees
        WHERE id = $1`

	err = h.db.QueryRow(context.Background(), query, id).Scan(
		&employee.ID,
		&employee.Name,
		&employee.Email,
		&employee.Role,
		&employee.Skills,
		&employee.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(employee)
}

func (h *EmployeeHandler) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	query := `
        SELECT id, name, email, role, skills, created_at,
        FROM employees`

	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var emp models.Employee
		err := rows.Scan(
			&emp.ID,
			&emp.Name,
			&emp.Email,
			&emp.Role,
			&emp.Skills,
			&emp.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		employees = append(employees, emp)
	}

	json.NewEncoder(w).Encode(employees)
}

func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var employee models.Employee
	if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
        UPDATE employees
        SET name = $1, email = $2, role = $3, skills = $4
        WHERE id = $5
        RETURNING id, name, email, role, skills, created_at`

	err = h.db.QueryRow(
		context.Background(),
		query,
		employee.Name,
		employee.Email,
		employee.Role,
		employee.Skills,
		id,
	).Scan(
		&employee.ID,
		&employee.Name,
		&employee.Email,
		&employee.Role,
		&employee.Skills,
		&employee.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(employee)
}
func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM employees WHERE id = $1`
	result, err := h.db.Exec(context.Background(), query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EmployeeHandler) GetEmployeeTasks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	query := `
        SELECT t.id, t.project_id, t.assigned_to, t.title, t.description, t.status, t.created_at
        FROM tasks t
        WHERE t.assigned_to = $1`

	rows, err := h.db.Query(context.Background(), query, employeeId)
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

	json.NewEncoder(w).Encode(tasks)
}
func (h *EmployeeHandler) GetEmployeeTasksByStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	status := vars["status"]
	if status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	query := `
        SELECT t.id, t.project_id, t.assigned_to, t.title, t.description, t.status, t.created_at
        FROM tasks t
        WHERE t.assigned_to = $1 AND t.status = $2`

	rows, err := h.db.Query(context.Background(), query, employeeId, status)
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

	json.NewEncoder(w).Encode(tasks)
}

func (h *EmployeeHandler) GetEmployeeProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	query := `
        SELECT p.id, p.name, p.description, p.lead_id, p.created_at
        FROM projects p
        JOIN employee_projects ep ON p.id = ep.project_id
        WHERE ep.employee_id = $1`

	rows, err := h.db.Query(context.Background(), query, employeeId)
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

	json.NewEncoder(w).Encode(projects)
}

func (h *EmployeeHandler) AssignEmployeeToProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeId, err := strconv.Atoi(vars["employeeId"])
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	projectId, err := strconv.Atoi(vars["projectId"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	query := `
        INSERT INTO employee_projects (employee_id, project_id)
        VALUES ($1, $2)
        ON CONFLICT (employee_id, project_id) DO NOTHING`

	_, err = h.db.Exec(context.Background(), query, employeeId, projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *EmployeeHandler) RemoveEmployeeFromProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeId, err := strconv.Atoi(vars["employeeId"])
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	projectId, err := strconv.Atoi(vars["projectId"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	query := `
        DELETE FROM employee_projects
        WHERE employee_id = $1 AND project_id = $2`

	result, err := h.db.Exec(context.Background(), query, employeeId, projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Assignment not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EmployeeHandler) GetEmployeesByProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	query := `
        SELECT e.id, e.name, e.email, e.role, e.created_at
        FROM employees e
        JOIN employee_projects ep ON e.id = ep.employee_id
        WHERE ep.project_id = $1`

	rows, err := h.db.Query(context.Background(), query, projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var emp models.Employee
		err := rows.Scan(
			&emp.ID,
			&emp.Name,
			&emp.Email,
			&emp.Role,
			&emp.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		employees = append(employees, emp)
	}

	json.NewEncoder(w).Encode(employees)
}
