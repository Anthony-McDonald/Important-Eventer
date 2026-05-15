package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

// Embed icons.
//
//go:embed icons/*
var embeddedIcons embed.FS

const staticPath = "static"
const iconDir = "icons"

// main is the application entry point that loads config and starts the HTTP server.
func main() {

	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	config = cfg

	fmt.Printf("fetching from %q\n", config.Calendar.CalendarURL)

	baseURL := fmt.Sprintf("%s:%d/%s/", config.Server.BaseURL, config.Server.Port, staticPath)

	nextHandler := MakeNextHandler(config.EmbeddingVectorURL, baseURL, embeddedIcons)

	http.HandleFunc("/next", nextHandler)

	subFS, err := fs.Sub(embeddedIcons, iconDir)
	if err != nil {
		panic(err)
	}

	http.Handle(fmt.Sprintf("/%s/%s/", staticPath, iconDir), http.StripPrefix(fmt.Sprintf("/%s/%s/", staticPath, iconDir),
		http.FileServer(http.FS(subFS)),
	))

	addr := fmt.Sprintf(":%d", config.Server.Port)

	fmt.Println("Server running on", addr)

	http.ListenAndServe(addr, nil)
}
