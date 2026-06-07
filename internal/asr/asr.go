package asr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/hybridgroup/yzma/pkg/llama"
	"github.com/hybridgroup/yzma/pkg/mtmd"
)

type Engine struct {
	ModelFile  string
	MMProjFile string
	LibPath    string

	model      llama.Model
	lctx       llama.Context
	mctx       mtmd.Context
	sampler    llama.Sampler
	vocab      llama.Vocab
	sampleRate int
	ready      bool
	mu         sync.Mutex
}

func New(modelFile, mmProjFile, libPath string) *Engine {
	return &Engine{
		ModelFile:  modelFile,
		MMProjFile: mmProjFile,
		LibPath:    libPath,
	}
}

func (e *Engine) Init() error {
	if err := llama.Load(e.LibPath); err != nil {
		return fmt.Errorf("loading llama library: %w", err)
	}
	if err := mtmd.Load(e.LibPath); err != nil {
		return fmt.Errorf("loading mtmd library: %w", err)
	}

	llama.LogSet(llama.LogSilent())
	mtmd.LogSet(llama.LogSilent())

	llama.Init()

	var err error
	e.model, err = llama.ModelLoadFromFile(e.ModelFile, llama.ModelDefaultParams())
	if err != nil {
		return fmt.Errorf("loading model: %w", err)
	}

	ctxParams := llama.ContextDefaultParams()
	ctxParams.NCtx = 4096
	ctxParams.NBatch = 2048

	e.lctx, err = llama.InitFromModel(e.model, ctxParams)
	if err != nil {
		llama.ModelFree(e.model)
		return fmt.Errorf("initializing context: %w", err)
	}

	e.vocab = llama.ModelGetVocab(e.model)

	sp := llama.DefaultSamplerParams()
	sp.Temp = 0.0
	sp.TopK = 1
	e.sampler = llama.NewSampler(e.model, llama.DefaultSamplers, sp)

	mctxParams := mtmd.ContextParamsDefault()
	e.mctx, err = mtmd.InitFromFile(e.MMProjFile, e.model, mctxParams)
	if err != nil {
		llama.Free(e.lctx)
		llama.ModelFree(e.model)
		return fmt.Errorf("initializing multimodal context: %w", err)
	}

	e.sampleRate = mtmd.GetAudioSampleRate(e.mctx)
	if e.sampleRate <= 0 {
		e.sampleRate = 16000
	}

	e.ready = true
	return nil
}

func (e *Engine) Close() {
	if e.sampler != 0 {
		llama.SamplerFree(e.sampler)
	}
	if e.mctx != 0 {
		mtmd.Free(e.mctx)
	}
	if e.lctx != 0 {
		llama.Free(e.lctx)
	}
	if e.model != 0 {
		llama.ModelFree(e.model)
	}
	llama.Close()
}

func (e *Engine) Transcribe(audioPath string) (string, error) {
	if !e.ready {
		return "", fmt.Errorf("engine not initialized")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	pcmPath, err := e.convertToPCM(audioPath)
	if err != nil {
		return "", fmt.Errorf("converting audio: %w", err)
	}
	defer os.Remove(pcmPath)

	samples, err := readPCMSamples(pcmPath)
	if err != nil {
		return "", fmt.Errorf("reading PCM samples: %w", err)
	}

	bitmap := mtmd.BitmapInitFromAudio(uint64(len(samples)), &samples[0])
	defer mtmd.BitmapFree(bitmap)

	prompt := e.buildPrompt()
	input := mtmd.NewInputText(prompt, true, true)
	output := mtmd.InputChunksInit()
	defer mtmd.InputChunksFree(output)

	if res := mtmd.Tokenize(e.mctx, output, input, []mtmd.Bitmap{bitmap}); res != 0 {
		return "", fmt.Errorf("tokenization failed: %d", res)
	}

	var n llama.Pos
	nBatch := llama.NBatch(e.lctx)
	if res := mtmd.HelperEvalChunks(e.mctx, e.lctx, output, 0, 0, int32(nBatch), true, &n); res != 0 {
		return "", fmt.Errorf("evaluation failed: %d", res)
	}

	var result string
	for i := 0; i < int(llama.MaxToken); i++ {
		token := llama.SamplerSample(e.sampler, e.lctx, -1)

		if llama.VocabIsEOG(e.vocab, token) {
			break
		}

		buf := make([]byte, 128)
		l := llama.TokenToPiece(e.vocab, token, buf, 0, true)
		result += string(buf[:l])

		batch := llama.BatchGetOne([]llama.Token{token})
		batch.Pos = &n

		llama.Decode(e.lctx, batch)
		n++
	}

	mem, err := llama.GetMemory(e.lctx)
	if err == nil && mem != 0 {
		llama.MemoryClear(mem, true)
	}

	return result, nil
}

func (e *Engine) buildPrompt() string {
	template := llama.ModelChatTemplate(e.model, "")
	if template != "" {
		messages := []llama.ChatMessage{
			llama.NewChatMessage("system", "You are a helpful assistant."),
			llama.NewChatMessage("user", mtmd.DefaultMarker()+"Transcribe the audio."),
		}
		buf := make([]byte, 16536)
		l := llama.ChatApplyTemplate(template, messages, true, buf)
		if l > 0 {
			return string(buf[:l])
		}
	}
	return mtmd.DefaultMarker() + "Transcribe the audio."
}

func (e *Engine) convertToPCM(audioPath string) (string, error) {
	pcmPath := audioPath + ".pcm"
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", audioPath,
		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", fmt.Sprintf("%d", e.sampleRate),
		"-ac", "1",
		pcmPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg: %s: %w", string(out), err)
	}
	return pcmPath, nil
}

func readPCMSamples(path string) ([]float32, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	samples := make([]float32, len(data)/4)
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &samples); err != nil {
		return nil, err
	}
	return samples, nil
}
