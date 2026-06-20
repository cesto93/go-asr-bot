package asr

type Engine interface {
	Init() error
	Close()
	Transcribe(pcm []float32) (string, error)
	TranscribeLang(pcm []float32, lang string) (string, error)
	SampleRate() int
	SetLanguage(lang string)
}
