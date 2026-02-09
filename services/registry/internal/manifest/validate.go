package manifest

import "fmt"

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
	if m.Audio != nil {
		if m.Audio.SampleRate <= 0 || m.Audio.Channels <= 0 {
			return fmt.Errorf("invalid audio spec")
		}
	}
	return nil
}
