package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "health check is accepted")
	})

	fmt.Println("Server is Runnning")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("HTTP Server 起動エラー", err)
	}
}
