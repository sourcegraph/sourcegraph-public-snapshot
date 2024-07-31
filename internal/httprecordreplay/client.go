package httprecordreplay

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func EnableHTTPRecordReplayFromEnv() {
	settings := recordReplaySettingsFromEnv()
	if settings.Mode == PassthroughMode {
		return
	}

	recordings := readRecordings(settings)
	if recordings.Settings.Mode == RecordMode {
		setupRecord(recordings)
	} else if recordings.Settings.Mode == ReplayMode {
		setupReplay(recordings)
	} else {
		panic(fmt.Sprintf("Invalid HTTP_RECORD_REPLAY_MODE: %s", recordings.Settings.Mode))
	}

}

func recordReplaySettingsFromEnv() RecordingSettings {
	return RecordingSettings{
		Mode:              recordReplayModeFromEnv(),
		RecordingFilePath: recordingFilePathFromEnv(),
	}
}

func recordingFilePathFromEnv() string {
	dir := os.Getenv("HTTP_RECORD_REPLAY_DIR")
	if dir == "" {
		// TODO: maybe default to a common directory?
		panic("HTTP_RECORD_REPLAY_DIR not set")
	}

	name := os.Getenv("HTTP_RECORD_REPLAY_NAME")
	if name == "" {
		panic("HTTP_RECORD_REPLAY_NAME not set")
	}

	return filepath.Join(dir, name+".recording.yaml")
}

type RecordingSettings struct {
	Mode              RecordReplayMode
	RecordingFilePath string
	Hasher            RequestHasher
}

// DefaultRequestHasher returns hashes a request based on its URL, method, and body.
// Note that it does not include headers.
func DefaultRequestHasher(req *http.Request) (string, error) {
	h := sha256.New()
	h.Write([]byte(req.URL.String()))
	h.Write([]byte(req.Method))
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil)), nil
}

type RequestHasher func(req *http.Request) (string, error)

func recordReplayModeFromEnv() RecordReplayMode {
	mode := os.Getenv("HTTP_RECORD_REPLAY_MODE")
	switch mode {
	case "", "passthrough":
		return PassthroughMode
	case "replay":
		return ReplayMode
	case "record":
		return RecordMode
	default:
		panic("Invalid HTTP_RECORD_REPLAY_MODE: " + mode)
	}
}

type RecordReplayMode string

const (
	ReplayMode      RecordReplayMode = "replay"
	RecordMode      RecordReplayMode = "record"
	PassthroughMode RecordReplayMode = "passthrough"
)

type Recordings struct {
	Settings                   RecordingSettings
	Recordings                 []*Recording
	shouldSaveRecordingsOnExit bool
}

func (r *Recordings) Find(requestHash string) *Recording {
	for _, rec := range r.Recordings {
		if requestHash == rec.Hash {
			return rec
		}
	}
	return nil
}

func (r *Recordings) HashOrPanic(req *http.Request) string {
	hash, err := r.Settings.Hasher(req)
	if err != nil {
		panic(fmt.Sprintf("Failed to hash request: %v", err))
	}
	return hash
}

func (r *Recordings) AddRecording(req *http.Request, hash string, res *http.Response) {
	r.Recordings = append(r.Recordings, &Recording{
		Hash:     hash,
		Request:  req,
		Response: res,
	})
	r.shouldSaveRecordingsOnExit = true
}

func (r *Recordings) OnExit() error {
	if !r.shouldSaveRecordingsOnExit {
		return nil
	}
	return r.Save()
}

func (r *Recordings) Save() error {
	filePath := filepath.Join(r.Settings.RecordingFilePath)
	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal recordings to YAML: %v", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write recordings to file %s: %v", filePath, err)
	}

	return nil
}

type Recording struct {
	Hash     string         `yaml:"hash"`
	Request  *http.Request  `yaml:"request"`
	Response *http.Response `yaml:"response"`
}

func readRecordings(settings RecordingSettings) *Recordings {
	rec := &Recordings{
		Settings:                   settings,
		Recordings:                 []*Recording{},
		shouldSaveRecordingsOnExit: false,
	}
	filePath := filepath.Join(settings.RecordingFilePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return rec
	}

	err = yaml.Unmarshal(data, rec)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal YAML from file %s: %v", filePath, err))
	}
	return rec
}

func setupRecord(recordings *Recordings) {
	httpcli.GlobalDoerMock = httpcli.DoerMock{
		DoFunc: func(underlyingDoer httpcli.DoerFunc, req *http.Request) (*http.Response, error) {
			requestHash := recordings.HashOrPanic(req)
			recording := recordings.Find(requestHash)
			if recording != nil {
				return recording.Response, nil
			}
			res, err := underlyingDoer(req)
			if err != nil {
				return nil, err
			}
			recordings.AddRecording(req, requestHash, res)

			recordings.Recordings = append(recordings.Recordings, &Recording{
				Hash:     requestHash,
				Request:  req,
				Response: res,
			})
			return underlyingDoer(req)
		},
	}
}

func setupReplay(recordings *Recordings) {
	httpcli.GlobalDoerMock = httpcli.DoerMock{
		DoFunc: func(underlyingDoer httpcli.DoerFunc, req *http.Request) (*http.Response, error) {
			requestHash := recordings.HashOrPanic(req)
			recording := recordings.Find(requestHash)
			if recording != nil {
				return recording.Response, nil
			}
			panic(fmt.Sprintf("No recording found for request with hash: %s", requestHash))
		},
	}

}
