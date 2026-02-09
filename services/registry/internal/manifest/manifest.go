package manifest

type Manifest struct {
	SchemaVersion string `json:"schema_version"`

	Model ModelSpec `json:"model"`

	Audio  *AudioSpec     `json:"audio,omitempty"`
	Labels map[string]int `json:"labels,omitempty"`

	Artifacts ArtifactsSpec `json:"artifacts"`
}

type ModelSpec struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
	Format    string `json:"format"`
	Precision string `json:"precision,omitempty"`
}

type AudioSpec struct {
	SampleRate    int `json:"sample_rate"`
	Channels      int `json:"channels"`
	MaxDurationMS int `json:"max_duration_ms"`
}

type ArtifactsSpec struct {
	Model            Artifact `json:"model"`
	Config           Artifact `json:"config"`
	FeatureExtractor Artifact `json:"feature_extractor"`
}

type Artifact struct {
	URL string `json:"url"`
}
