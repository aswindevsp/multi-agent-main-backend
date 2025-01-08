package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"nstorm.com/main-backend/handlers"
	"nstorm.com/main-backend/ollama"
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

	// Initialize Ollama client
	ollamaClient := ollama.NewClient("http://localhost:11434")
	employeeHandler := handlers.NewEmployeeHandler(conn)
	projectHandler := handlers.NewProjectHandler(conn)
	llmHandler := handlers.NewLLMHandler(ollamaClient, conn)

	router := mux.NewRouter()

	// Employee routes
	router.HandleFunc("/api/employees/tasks", employeeHandler.GetEmployeeTasks).Methods("GET")
	router.HandleFunc("/api/employees/assign-project", employeeHandler.AssignToProject).Methods("POST")
	router.HandleFunc("/api/employees/remove-from-project", employeeHandler.RemoveFromProject).Methods("POST")
	router.HandleFunc("/api/employees/assign-task", employeeHandler.AssignTask).Methods("POST")
	router.HandleFunc("/api/employees/complete-task", employeeHandler.CompleteTask).Methods("POST")

	// Project routes
	router.HandleFunc("/api/projects", projectHandler.CreateProject).Methods("POST")
	router.HandleFunc("/api/projects/{id}", projectHandler.DeleteProject).Methods("DELETE")
	router.HandleFunc("/api/projects/{id}", projectHandler.UpdateProject).Methods("PUT")
	router.HandleFunc("/api/projects/{id}/tasks", projectHandler.GetProjectTasks).Methods("GET")

	// LLM routes
	router.HandleFunc("/api/project/plan", llmHandler.CreateProjectPlan).Methods("POST")

	fmt.Println("Server starting on port 8888...")
	http.ListenAndServe(":8888", router)
}
