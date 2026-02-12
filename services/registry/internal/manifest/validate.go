package manifest

import (
	"fmt"
)

func (m *Manifest) Validate() error {
	if m.SchemaVersion == "" {
		return fmt.Errorf("schema_version required")
	}
	if m.Model.Name == "" || m.Model.Version == "" {
		return fmt.Errorf("model name/version required")
	}

	if len(m.Artifacts) == 0 {
		return fmt.Errorf("artifacts required")
	}

	for name, artifact := range m.Artifacts {
		if artifact.URL == "" {
			return fmt.Errorf("artifact '%s.url' required", name)
		}
		if artifact.SHA256 == "" {
			return fmt.Errorf("artifact '%s.sha256' required", name)
		}
		if artifact.Type == "" {
			return fmt.Errorf("artifact '%s.type' required", name)
		}
	}

	if _, ok := m.Artifacts["model"]; !ok {
		return fmt.Errorf("artifact 'model' required")
	}

	if m.Audio != nil {
		if m.Audio.SampleRate <= 0 {
			return fmt.Errorf("audio.sample_rate must be > 0")
		}
		if m.Audio.Channels <= 0 {
			return fmt.Errorf("audio.channels must be > 0")
		}
	}

	if m.Labels != nil && len(m.Labels) == 0 {
		return fmt.Errorf("labels present but empty")
	}

	return nil
}
