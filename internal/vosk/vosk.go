package vosk

/*
#cgo linux LDFLAGS: -ldl -lpthread
#cgo darwin LDFLAGS: -ldl -lpthread
#cgo windows LDFLAGS: -lkernel32
#include <stdlib.h>
#include <string.h>

#ifdef _WIN32
#include <windows.h>
#else
#include <dlfcn.h>
#endif

typedef struct VoskModel VoskModel;
typedef struct VoskRecognizer VoskRecognizer;

typedef void (*set_log_level_fn)(int);
typedef VoskModel* (*model_new_fn)(const char *);
typedef void (*model_free_fn)(VoskModel *);
typedef VoskRecognizer* (*recognizer_new_fn)(VoskModel *, float);
typedef void (*recognizer_free_fn)(VoskRecognizer *);
typedef void (*recognizer_set_words_fn)(VoskRecognizer *, int);
typedef int (*recognizer_accept_waveform_fn)(VoskRecognizer *, const char *, int);
typedef const char* (*recognizer_final_result_fn)(VoskRecognizer *);

static void *vosk_handle = NULL;
static char vosk_error[512];
static set_log_level_fn p_set_log_level;
static model_new_fn p_model_new;
static model_free_fn p_model_free;
static recognizer_new_fn p_recognizer_new;
static recognizer_free_fn p_recognizer_free;
static recognizer_set_words_fn p_recognizer_set_words;
static recognizer_accept_waveform_fn p_recognizer_accept_waveform;
static recognizer_final_result_fn p_recognizer_final_result;

static void set_error(const char *message) {
	strncpy(vosk_error, message, sizeof(vosk_error) - 1);
	vosk_error[sizeof(vosk_error) - 1] = '\0';
}

static void *load_symbol(const char *name) {
#ifdef _WIN32
	return (void *)GetProcAddress((HMODULE)vosk_handle, name);
#else
	return dlsym(vosk_handle, name);
#endif
}

static int vosk_load_library(const char *path) {
	if (vosk_handle != NULL) return 1;
#ifdef _WIN32
	vosk_handle = (void *)LoadLibraryA(path);
	if (vosk_handle == NULL) { set_error("LoadLibrary failed for Vosk library"); return 0; }
#else
	vosk_handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
	if (vosk_handle == NULL) { set_error(dlerror()); return 0; }
#endif
#define LOAD(name) do { p_##name = (name##_fn)load_symbol("vosk_" #name); if (p_##name == NULL) { set_error("Vosk library is missing symbol vosk_" #name); return 0; } } while (0)
	LOAD(set_log_level);
	LOAD(model_new);
	LOAD(model_free);
	LOAD(recognizer_new);
	LOAD(recognizer_free);
	LOAD(recognizer_set_words);
	LOAD(recognizer_accept_waveform);
	LOAD(recognizer_final_result);
#undef LOAD
	return 1;
}

static const char *vosk_last_error() { return vosk_error; }
static void call_set_log_level(int level) { p_set_log_level(level); }
static VoskModel *call_model_new(const char *path) { return p_model_new(path); }
static void call_model_free(VoskModel *model) { p_model_free(model); }
static VoskRecognizer *call_recognizer_new(VoskModel *model, float rate) { return p_recognizer_new(model, rate); }
static void call_recognizer_free(VoskRecognizer *recognizer) { p_recognizer_free(recognizer); }
static void call_recognizer_set_words(VoskRecognizer *recognizer, int enabled) { p_recognizer_set_words(recognizer, enabled); }
static int call_recognizer_accept_waveform(VoskRecognizer *recognizer, const char *data, int length) { return p_recognizer_accept_waveform(recognizer, data, length); }
static const char *call_recognizer_final_result(VoskRecognizer *recognizer) { return p_recognizer_final_result(recognizer); }
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"
)

var loadMu sync.Mutex

func LoadLibrary(path string) error {
	loadMu.Lock()
	defer loadMu.Unlock()

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	if C.vosk_load_library(cpath) == 0 {
		return fmt.Errorf("load Vosk library %q: %s", path, C.GoString(C.vosk_last_error()))
	}
	return nil
}

func SetLogLevel(level int) { C.call_set_log_level(C.int(level)) }

type Model struct{ native *C.VoskModel }

func NewModel(path string) (*Model, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	native := C.call_model_new(cpath)
	if native == nil {
		return nil, fmt.Errorf("Vosk could not load model from %q", path)
	}
	return &Model{native: native}, nil
}

func (m *Model) Free() {
	if m != nil && m.native != nil {
		C.call_model_free(m.native)
		m.native = nil
	}
}

type Recognizer struct{ native *C.VoskRecognizer }

func NewRecognizer(model *Model, sampleRate float64) (*Recognizer, error) {
	if model == nil || model.native == nil {
		return nil, fmt.Errorf("Vosk model is not initialized")
	}
	native := C.call_recognizer_new(model.native, C.float(sampleRate))
	if native == nil {
		return nil, fmt.Errorf("Vosk could not create recognizer")
	}
	return &Recognizer{native: native}, nil
}

func (r *Recognizer) Free() {
	if r != nil && r.native != nil {
		C.call_recognizer_free(r.native)
		r.native = nil
	}
}

func (r *Recognizer) SetWords(enabled bool) {
	value := C.int(0)
	if enabled {
		value = 1
	}
	C.call_recognizer_set_words(r.native, value)
}

func (r *Recognizer) AcceptWaveform(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	return C.call_recognizer_accept_waveform(r.native, (*C.char)(unsafe.Pointer(&data[0])), C.int(len(data))) != 0
}

func (r *Recognizer) FinalResult() string {
	return C.GoString(C.call_recognizer_final_result(r.native))
}
