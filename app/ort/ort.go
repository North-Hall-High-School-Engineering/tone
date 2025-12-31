package ort

import (
	"app/assets"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

const ONNX_VERSION = "1.23.2"

// Extracts embedded onxxruntime(.dll | .dylib | .so) to concrete location, returns concrete path and error
func ExtractEmbeddedOrt() (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Library name in embedding
	var lib string
	switch goos {
	case "darwin":
		lib = "libonnxruntime." + ONNX_VERSION + ".dylib"
	case "linux":
		lib = "libonnxruntime.so." + ONNX_VERSION
	case "windows":
		lib = "onnxruntime.dll"
	default:
		return "", fmt.Errorf("unsupported os %s", goos)
	}

	// Source path of embedded library
	src := path.Join("onnxruntime", goos, goarch, lib)

	data, err := assets.FS.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded file %s, %v", src, err.Error())
	}

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, filepath.Base(lib))

	err = os.WriteFile(tmpFile, data, 0755)
	if err != nil {
		return "", err
	}

	return tmpFile, nil
}
