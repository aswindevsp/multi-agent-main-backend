package main

import (
	"context"
	"fmt"
	"net/http"
	"nstorm.com/main-backend/database"
	"os"

	"github.com/gorilla/mux"
	"nstorm.com/main-backend/handlers"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	conn, err := database.NewConnection()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	employeeHandler := handlers.NewEmployeeHandler(conn)
	projectHandler := handlers.NewProjectHandler(conn)
	taskHandler := handlers.NewTaskHandler(conn)

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

	handler := corsMiddleware(router)

	fmt.Println("Server starting on port 8888...")
	http.ListenAndServe(":8888", handler)

}
