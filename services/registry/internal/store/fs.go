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
	path := filepath.Join(s.Path, name, version+".json")
	log.Println(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m manifest.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, m.Validate()
}
