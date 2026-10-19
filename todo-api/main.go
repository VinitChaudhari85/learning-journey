package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
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
	if err := decodeJSON(r, &newTodo); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if newTodo.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i, t := range todos {
		if t.ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	writeError(w, http.StatusNotFound, "todo not found")
}

type updateInput struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	var input updateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Title == nil && input.Done == nil {
		writeError(w, http.StatusBadRequest, "nothing to update")
		return
	}
	if input.Title != nil && *input.Title == "" {
		writeError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i := range todos {
		if todos[i].ID == id {
			if input.Title != nil {
				todos[i].Title = *input.Title
			}
			if input.Done != nil {
				todos[i].Done = *input.Done
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(todos[i])
			return
		}
	}

	writeError(w, http.StatusNotFound, "todo not found")
}

func countTodos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprint(w, len(todos))
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func main() {
	http.HandleFunc("GET /todos", getTodos)
	http.HandleFunc("POST /todos", createTodo)
	http.HandleFunc("DELETE /todos/{id}", deleteTodo)
	http.HandleFunc("PUT /todos/{id}", updateTodo)
	http.HandleFunc("GET /count", countTodos)
	http.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		fmt.Fprintln(w, "finished slow request")
	})

	srv := &http.Server{Addr: ":8080"}

	go func() {
		fmt.Println("Listening on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("forced shutdown:", err)
	}
	fmt.Println("Server stopped")
}
