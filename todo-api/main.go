package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var todos = []Todo{
	{ID: 1, Title: "Learn Go", Done: true},
	{ID: 2, Title: "Build a REST API", Done: false},
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func main() {
	http.HandleFunc("GET /todos", getTodos)

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
