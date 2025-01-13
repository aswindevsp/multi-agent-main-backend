package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"nstorm.com/main-backend/handlers"
)

func main() {
	// Database connection
	connectionUrl := "postgres://postgres:example@localhost:5432/multiagent"
	conn, err := pgx.Connect(context.Background(), connectionUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// ollamaClient := ollama.NewClient("http://localhost:11434")
	employeeHandler := handlers.NewEmployeeHandler(conn)
	projectHandler := handlers.NewProjectHandler(conn)
	taskHandler := handlers.NewTaskHandler(conn)
	// llmHandler := handlers.NewLLMHandler(ollamaClient, conn)

	router := mux.NewRouter()

	router.HandleFunc("/employees", employeeHandler.GetAllEmployees).Methods("GET")
	router.HandleFunc("/employees", employeeHandler.CreateEmployee).Methods("POST")
	router.HandleFunc("/employees/{id}", employeeHandler.GetEmployeeById).Methods("GET")
	router.HandleFunc("/employees/{id}", employeeHandler.UpdateEmployee).Methods("PUT")
	router.HandleFunc("/employees/{id}", employeeHandler.DeleteEmployee).Methods("DELETE")

	router.HandleFunc("/employees/{id}/tasks", employeeHandler.GetEmployeeTasks).Methods("GET")
	router.HandleFunc("/employees/{id}/tasks/{status}", employeeHandler.GetEmployeeTasksByStatus).Methods("GET")
	router.HandleFunc("/employees/{id}/projects", employeeHandler.GetEmployeeProjects).Methods("GET")
	router.HandleFunc("/employees/{employeeId}/projects/{projectId}", employeeHandler.AssignEmployeeToProject).Methods("POST")
	router.HandleFunc("/employees/{employeeId}/projects/{projectId}", employeeHandler.RemoveEmployeeFromProject).Methods("DELETE")
	router.HandleFunc("/projects/{id}/employees", employeeHandler.GetEmployeesByProject).Methods("GET")

	router.HandleFunc("/projects", projectHandler.GetAllProjects).Methods("GET")
	router.HandleFunc("/projects", projectHandler.CreateProject).Methods("POST")
	router.HandleFunc("/projects/{id}", projectHandler.GetProjectByID).Methods("GET")
	router.HandleFunc("/projects/{id}", projectHandler.UpdateProject).Methods("PUT")
	router.HandleFunc("/projects/{id}", projectHandler.DeleteProject).Methods("DELETE")

	router.HandleFunc("/tasks", taskHandler.GetAllTasks).Methods("GET")
	router.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	router.HandleFunc("/tasks/{id}", taskHandler.GetTaskByID).Methods("GET")
	router.HandleFunc("/tasks/{id}", taskHandler.UpdateTask).Methods("PUT")
	router.HandleFunc("/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")
	router.HandleFunc("/projects/{id}/generate-tasks", projectHandler.GenerateAndAssignTasks).Methods("POST")

	fmt.Println("Server starting on port 8888...")
	http.ListenAndServe(":8888", router)

}
