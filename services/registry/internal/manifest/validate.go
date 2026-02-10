package manifest

import "fmt"

func (m *Manifest) Validate() error {
	if m.SchemaVersion == "" {
		return fmt.Errorf("schema_version required")
	}
	if m.Model.Name == "" || m.Model.Version == "" {
		return fmt.Errorf("model name/version required")
	}

	if m.Artifacts.Model.URL == "" {
		return fmt.Errorf("artifact 'model.url' required")
	}
	// if m.Artifacts.Config.URL == "" {
	// 	return fmt.Errorf("artifact 'config.url' required")
	// }
	if m.Artifacts.FeatureExtractor.URL == "" {
		return fmt.Errorf("artifact 'feature_extractor.url' required")
	}

	if m.Audio != nil {
		if m.Audio.SampleRate <= 0 {
			return fmt.Errorf("audio.sample_rate must be > 0")
		}
		if m.Audio.Channels <= 0 {
			return fmt.Errorf("audio.channels must be > 0")
		}
		if m.Audio.MaxDurationMS <= 0 {
			return fmt.Errorf("audio.max_duration_ms must be > 0")
		}
	}

	if m.Labels != nil && len(m.Labels) == 0 {
		return fmt.Errorf("labels present but empty")
	}

	return nil
}
