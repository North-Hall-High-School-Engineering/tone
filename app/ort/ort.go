package ort

import (
	"app/assets"
	"fmt"
	"os"
	"path"
	"runtime"
)

const ONNX_VERSION = "1.23.2"

// Extracts embedded onxxruntime(.dll | .dylib | .so) and model.onnx to concrete location, returns lib path, model path, and error
func ExtractEmbeddedOrt() (string, string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var lib string
	switch goos {
	case "darwin":
		lib = "libonnxruntime." + ONNX_VERSION + ".dylib"
	case "linux":
		lib = "libonnxruntime.so." + ONNX_VERSION
	case "windows":
		lib = "onnxruntime.dll"
	default:
		return "", "", fmt.Errorf("unsupported os %s", goos)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}

	runtimeRoot := path.Join(configDir, "tone", "runtime")

	libRelPath := path.Join("onnxruntime", goos, goarch, lib)
	modelRelPath := path.Join("models", "tone.onnx")
	modelVerRelPath := path.Join("models", "tone.onnx.version")

	libDest := path.Join(runtimeRoot, libRelPath)
	modelDest := path.Join(runtimeRoot, modelRelPath)
	modelVerDest := path.Join(runtimeRoot, modelVerRelPath)

	upToDate, err := isUpToDate(modelVerRelPath, modelVerDest)
	if err != nil {
		return "", "", err
	}

	if upToDate {
		return libDest, modelDest, nil
	}

	files := map[string]string{
		libRelPath:      libDest,
		modelRelPath:    modelDest,
		modelVerRelPath: modelVerDest,
	}

	for src, dest := range files {
		data, err := assets.FS.ReadFile(src)
		if err != nil {
			return "", "", fmt.Errorf("read embedded %s: %w", src, err)
		}

		if err := os.MkdirAll(path.Dir(dest), 0755); err != nil {
			return "", "", err
		}

		if err := os.WriteFile(dest, data, 0644); err != nil {
			return "", "", fmt.Errorf("write %s: %w", dest, err)
		}
	}

	return libDest, modelDest, nil
}

func isUpToDate(embeddedVerPath, extractedVerPath string) (bool, error) {
	embeddedVer, err := assets.FS.ReadFile(embeddedVerPath)
	if err != nil {
		return false, err
	}

	existingVer, err := os.ReadFile(extractedVerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return string(embeddedVer) == string(existingVer), nil
}
