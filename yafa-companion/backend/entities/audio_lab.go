package entities

// Assets are generated/converted on the PC; firmware receives only PCM samples.
type LabSpeech struct {
	Audio              []byte `json:"audio"`
	Format             string `json:"format"`
	PCM                []byte `json:"pcm"`
	SampleRate         int    `json:"sample_rate"`
	TruncatedAtSeconds int    `json:"maximum_seconds"`
}
type LabTranscript struct {
	Text  string `json:"text"`
	Model string `json:"model"`
}
