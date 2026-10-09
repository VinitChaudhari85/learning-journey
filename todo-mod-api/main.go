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

// ---------- STORE: all data + all locking lives here ----------

type Store struct {
	mu     sync.Mutex
	nextID int
	todos  []Todo
}

func NewStore() *Store {
	return &Store{
		nextID: 3,
		todos: []Todo{
			{ID: 1, Title: "Learn Go", Done: true},
			{ID: 2, Title: "Build a REST API", Done: false},
		},
	}
}

func (s *Store) List() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Todo, len(s.todos))
	copy(out, s.todos)
	return out
}

func (s *Store) Create(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Todo{ID: s.nextID, Title: title}
	s.nextID++
	s.todos = append(s.todos, t)
	return t
}

func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.todos {
		if t.ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) Update(id int, title string, done bool) (Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.todos {
		if s.todos[i].ID == id {
			s.todos[i].Title = title
			s.todos[i].Done = done
			return s.todos[i], true
		}
	}
	return Todo{}, false
}

// ---------- HANDLERS: only HTTP, no locks, no direct todos access ----------

type API struct {
	store *Store
}

func (a *API) getTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a.store.List())
}

func (a *API) countTodos(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, len(a.store.List())) // no lock, and no race
}

func (a *API) createTodo(w http.ResponseWriter, r *http.Request) {
	var input Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if input.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	newTodo := a.store.Create(input.Title)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

func (a *API) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be a number", http.StatusBadRequest)
		return
	}

	if !a.store.Delete(id) {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be a number", http.StatusBadRequest)
		return
	}

	var input Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	updated, found := a.store.Update(id, input.Title, input.Done)
	if !found {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func main() {
	api := &API{store: NewStore()}

	http.HandleFunc("GET /todos", api.getTodos)
	http.HandleFunc("GET /count", api.countTodos)
	http.HandleFunc("POST /todos", api.createTodo)
	http.HandleFunc("DELETE /todos/{id}", api.deleteTodo)
	http.HandleFunc("PUT /todos/{id}", api.updateTodo)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
