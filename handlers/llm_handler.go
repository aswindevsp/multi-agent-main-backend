package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"nstorm.com/main-backend/ollama"
)

type LLMHandler struct {
	client *ollama.Client
	db     *pgx.Conn
}

type ProjectRequest struct {
	Requirements string   `json:"requirements"`
	Employees    []string `json:"employees"`
}

type ProjectTask struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Duration     string   `json:"duration"`
	Assignees    []string `json:"assignees"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type ProjectPlan struct {
	Tasks    []ProjectTask `json:"tasks"`
	Timeline string        `json:"timeline"`
}

func NewLLMHandler(client *ollama.Client, db *pgx.Conn) *LLMHandler {
	return &LLMHandler{
		client: client,
		db:     db,
	}
}

func (h *LLMHandler) CreateProjectPlan(w http.ResponseWriter, r *http.Request) {
	var request ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(`Create a detailed project plan based on these requirements:
    %s
    
    Available team members: %s
    
    Format the response as JSON with this structure:
    {
        "tasks": [
            {
                "title": "Task name",
                "description": "Detailed description",
                "duration": "Estimated duration",
                "assignees": ["Team member names"],
                "dependencies": ["Dependent task titles"]
            }
        ],
        "timeline": "Total project timeline estimate"
    }
    
    Consider dependencies between tasks and team member expertise.`,
		request.Requirements, strings.Join(request.Employees, ", "))

	response, err := h.client.Generate(context.Background(), "llama2", prompt)
	if err != nil {
		http.Error(w, "Failed to generate project plan", http.StatusInternalServerError)
		return
	}

	jsonStr := extractJSON(response.Content)
	var plan ProjectPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		http.Error(w, "Failed to parse project plan", http.StatusInternalServerError)
		return
	}

	// Validate and clean up the plan
	if len(plan.Tasks) == 0 {
		http.Error(w, "No tasks generated in project plan", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func extractJSON(response string) string {
	// Find the first '{' and last '}'
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 || end < start {
		return "{}" // Return empty JSON if no valid JSON found
	}

	// Extract the JSON substring
	jsonStr := response[start : end+1]

	// Clean up common LLM formatting issues
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\\", "")

	// Remove any markdown code block markers
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimSuffix(jsonStr, "```")

	return strings.TrimSpace(jsonStr)
}
