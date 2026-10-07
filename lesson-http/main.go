package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	// Reading the request
	fmt.Println("The request method: ", r.Method)
	fmt.Println("The URL path: ", r.URL.Path)
	fmt.Println("User-Agent: ", r.Header.Get("User-Agent"))
	// Writing the response
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	fmt.Fprintln(w, "Hello from Go!")
}

func main() {
	http.HandleFunc("GET /hello", hello)
	fmt.Println("Listening on port :8080")
	http.ListenAndServe(":8080", nil)
}
