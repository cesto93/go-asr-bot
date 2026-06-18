package asr

type Engine interface {
	Init() error
	Close()
	Transcribe(pcm []float32) (string, error)
	SampleRate() int
}
