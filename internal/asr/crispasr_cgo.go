//go:build cgo

package asr

/*
#cgo LDFLAGS: -L${SRCDIR}/../../lib/crispasr/build/src -lcrispasr
#cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/../../lib/crispasr/build/src

#include <stdlib.h>

typedef struct CrispasrSession CrispasrSession;
typedef struct crispasr_session_result crispasr_session_result;

CrispasrSession* crispasr_session_open(const char* model_path, int n_threads);
void             crispasr_session_close(CrispasrSession* s);

crispasr_session_result* crispasr_session_transcribe(CrispasrSession* s, const float* pcm, int n_samples);
int          crispasr_session_result_n_segments(crispasr_session_result* r);
const char*  crispasr_session_result_segment_text(crispasr_session_result* r, int i);
long long    crispasr_session_result_segment_t0(crispasr_session_result* r, int i);
long long    crispasr_session_result_segment_t1(crispasr_session_result* r, int i);
void         crispasr_session_result_free(crispasr_session_result* r);

int crispasr_session_set_max_new_tokens(CrispasrSession* s, int max_new_tokens);
int crispasr_session_set_temperature(CrispasrSession* s, float temperature, unsigned long long seed);
int crispasr_session_set_source_language(CrispasrSession* s, const char* lang);
crispasr_session_result* crispasr_session_transcribe_lang(CrispasrSession* s, const float* pcm, int n_samples, const char* language);
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

type crispasrEngine struct {
	session  *C.CrispasrSession
	language string
}

func newCrispasr(modelPath string, threads int, lang string) (*crispasrEngine, error) {
	cpath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cpath))

	h := C.crispasr_session_open(cpath, C.int(threads))
	if h == nil {
		return nil, fmt.Errorf("crispasr_session_open: failed to open %s", modelPath)
	}

	engine := &crispasrEngine{session: h, language: lang}

	C.crispasr_session_set_max_new_tokens(h, C.int(256))
	C.crispasr_session_set_temperature(h, C.float(0), C.ulonglong(0))

	if lang != "" {
		clang := C.CString(lang)
		C.crispasr_session_set_source_language(h, clang)
		C.free(unsafe.Pointer(clang))
	}

	return engine, nil
}

func (e *crispasrEngine) Init() error {
	if e.session == nil {
		return fmt.Errorf("crispasr engine not created")
	}
	return nil
}

func (e *crispasrEngine) Close() {
	if e.session != nil {
		C.crispasr_session_close(e.session)
		e.session = nil
	}
}

func (e *crispasrEngine) SampleRate() int {
	return 16000
}

func (e *crispasrEngine) Transcribe(pcm []float32) (string, error) {
	return e.TranscribeLang(pcm, e.language)
}

func (e *crispasrEngine) TranscribeLang(pcm []float32, lang string) (string, error) {
	if e.session == nil {
		return "", fmt.Errorf("engine not initialized")
	}

	pcmPtr := (*C.float)(unsafe.Pointer(nil))
	if len(pcm) > 0 {
		pcmPtr = (*C.float)(unsafe.Pointer(&pcm[0]))
	}

	var r *C.crispasr_session_result
	if lang != "" {
		clang := C.CString(lang)
		r = C.crispasr_session_transcribe_lang(e.session, pcmPtr, C.int(len(pcm)), clang)
		C.free(unsafe.Pointer(clang))
	} else {
		r = C.crispasr_session_transcribe(e.session, pcmPtr, C.int(len(pcm)))
	}
	if r == nil {
		return "", fmt.Errorf("transcription failed")
	}
	defer C.crispasr_session_result_free(r)

	nSegs := int(C.crispasr_session_result_n_segments(r))
	var texts []string
	for i := 0; i < nSegs; i++ {
		text := C.GoString(C.crispasr_session_result_segment_text(r, C.int(i)))
		texts = append(texts, text)
	}

	return strings.Join(texts, " "), nil
}

func init() {
	backends["crispasr"] = func(modelPath string, threads int, lang string) (Engine, error) {
		return newCrispasr(modelPath, threads, lang)
	}
}
