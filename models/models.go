package models

import "time"

type EmployeeRole string

const (
	RoleProjectManager EmployeeRole = "PROJECT_MANAGER"
	RoleDeveloper      EmployeeRole = "DEVELOPER"
)

type Employee struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Role      EmployeeRole `json:"role"`
	CreatedAt time.Time    `json:"created_at"`
	Projects  []Project    `json:"projects,omitempty"`
	Tasks     []Task       `json:"tasks,omitempty`
}

type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LeadID      int       `json:"lead_id"`
	CreatedAt   time.Time `json:"created_at"`
	Tasks       []Task    `json:"tasks,omitempty"`
}

type Task struct {
	ID          int       `json:"id"`
	ProjectID   int       `json:"project_id"`
	AssignedTo  int       `json:"assigned_to"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
