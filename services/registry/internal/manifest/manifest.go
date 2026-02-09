package manifest

type Manifest struct {
	SchemaVersion string `json:"schema_version"`

	Model struct {
		Name      string `json:"name"`
		Version   string `json:"version"`
		SHA256    string `json:"sha256"`
		SizeBytes int64  `json:"size_bytes"`
		Format    string `json:"format"`
		Precision string `json:"precision,omitempty"`
	} `json:"model"`

	Audio  *AudioSpec     `json:"audio,omitempty"`
	Labels map[string]int `json:"labels,omitempty"`

	Artifacts map[string]string `json:"artifacts"`
}

type AudioSpec struct {
	SampleRate    int `json:"sample_rate"`
	Channels      int `json:"channels"`
	MaxDurationMS int `json:"max_duration_ms"`
}
