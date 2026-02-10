package store

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/North-Hall-High-School-Engineering/tone/services/registry/internal/manifest"
)

type FS struct {
	Path string
}

func (s *FS) Load(name, version string) (*manifest.Manifest, error) {
	// Print the base path for debugging
	log.Println("Base path:", s.Path)

	// List all files in the base path
	files, err := os.ReadDir(s.Path)
	if err != nil {
		log.Println("Error reading base path:", err)
	} else {
		log.Println("Files in base path:")
		for _, f := range files {
			log.Println(" -", f.Name())
		}
	}

	// Construct the manifest file path
	path := filepath.Join(s.Path, name+"-"+version+".json")
	log.Println("Manifest path:", path)

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	var m manifest.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return &m, m.Validate()
}

// func (s *FS) Load(name, version string) (*manifest.Manifest, error) {
// 	path := filepath.Join(s.Path, name+"-"+version+".json")
// 	log.Println(path)
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var m manifest.Manifest
// 	if err := json.Unmarshal(data, &m); err != nil {
// 		return nil, err
// 	}
// 	return &m, m.Validate()
// }
