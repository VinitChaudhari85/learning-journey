package main 

import (
	"fmt"
	"net/http"
)

func main(){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		fmt.Fprintln(w, "Hello from my Go server, Day 2!")
	})

	fmt.Println("Server running on http://localhost:8000")
	http.ListenAndServe(".8000", nil)
}