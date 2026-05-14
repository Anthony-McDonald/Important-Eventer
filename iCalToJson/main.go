package main

import (
	"fmt"
	"net/http"
)

// main is the application entry point that loads config and starts the HTTP server.
func main() {

	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	config = cfg

	fmt.Printf("fetching from %q\n", config.Calendar.URL)

	http.HandleFunc("/next", handler)

	addr := fmt.Sprintf(":%d", config.Server.Port)

	fmt.Println("Server running on", addr)

	http.ListenAndServe(addr, nil)
}
