package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var (
	mu     sync.Mutex
	nextID = 3
	todos  = []Todo{
		{ID: 1, Title: "Learn Go", Done: true},
		{ID: 2, Title: "Build a REST API", Done: false},
	}
)

func getTodos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
	mu.Unlock()
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo
	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if newTodo.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	newTodo.ID = nextID
	nextID++
	todos = append(todos, newTodo)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	// 1. READ the id from the URL
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id must be a number", http.StatusBadRequest)
		return
	}

	// 2. FIND it and remove it (locked, like before)
	mu.Lock()
	defer mu.Unlock()

	for i, t := range todos {
		if t.ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// 3. Not found
	http.Error(w, "todo not found", http.StatusNotFound)
}

func main() {
	http.HandleFunc("GET /todos", getTodos)
	http.HandleFunc("POST /todos", createTodo)
	http.HandleFunc("DELETE /todos/{id}", deleteTodo)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
