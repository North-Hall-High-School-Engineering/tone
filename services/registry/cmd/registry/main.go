package main

import (
	"log"
	"net/http"
	"os"

	"github.com/North-Hall-High-School-Engineering/tone/services/registry/internal/api"
	"github.com/North-Hall-High-School-Engineering/tone/services/registry/internal/store"
)

func main() {
	basePath := os.Getenv("MODEL_MANIFEST_PATH")
	if basePath == "" {
		basePath = "/manifests"
	}

	h := &api.Handler{
		Store: &store.FS{Path: basePath},
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":8080", api.Routes(h)); err != nil {
		log.Fatal(err)
	}

}
