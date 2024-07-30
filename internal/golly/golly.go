package golly

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GollyRecordingMode string

const (
	RecordingModeReplay      GollyRecordingMode = "replay"
	RecordingModeRecord      GollyRecordingMode = "record"
	RecordingModePassthrough GollyRecordingMode = "passthrough"
)

type Golly struct {
	T                             *testing.T
	Credentials                   []TestingCredentials
	RecordingFilePath             string
	RecordingName                 string
	RecordingMode                 GollyRecordingMode
	Doer                          httpcli.Doer
	Recordings                    []*yamlRecording
	shouldSaveRecordingsOnCleanup bool
	Hasher                        RequestHasher
	ResponseHeaderMatcher         func(header string, value string) bool
	RequestHeaderMatcher          func(header string, value string) bool
}

func (g *Golly) DotcomCredentials() *TestingCredentials {
	for _, cred := range g.Credentials {
		if cred.Endpoint == "https://sourcegraph.com" {
			return &cred
		}
	}
	g.T.Fatalf("Could not find dotcom access token")
	return nil
}

// DefaultRequestHasher returns hashes a request based on its URL, method, and body.
// Note that it does not include headers.
func DefaultRequestHasher(t *testing.T) RequestHasher {
	return func(req *http.Request) (string, error) {
		h := sha256.New()
		h.Write([]byte(req.URL.String()))
		h.Write([]byte(req.Method))
		body := readyBodyIntoMemory(t, req)
		h.Write(body)
		return hex.EncodeToString(h.Sum(nil)), nil
	}
}

type RequestHasher func(req *http.Request) (string, error)

var _ httpcli.Doer = (*Golly)(nil)

func (g *Golly) Do(r *http.Request) (*http.Response, error) {
	if g.RecordingMode == RecordingModeReplay {
		return g.replay(r)
	}
	if g.RecordingMode == RecordingModeRecord {
		return g.record(r)
	}
	return g.passthrough(r)
}

func (g *Golly) record(r *http.Request) (*http.Response, error) {
	requestHash := g.HashOrPanic(r)
	recording := g.Find(requestHash)
	if recording != nil {
		return recording.Response.HttpResponse(), nil
	}
	res, err := g.passthrough(r)
	if err != nil {
		return nil, err
	}
	g.AddRecording(r, requestHash, res)

	freshRecording := g.newYamlRecording(requestHash, r, res)
	g.Recordings = append(g.Recordings, freshRecording)
	return freshRecording.Response.HttpResponse(), nil
}

func (g *Golly) replay(r *http.Request) (*http.Response, error) {
	requestHash := g.HashOrPanic(r)
	recording := g.Find(requestHash)
	if recording != nil {
		return recording.Response.HttpResponse(), nil
	}
	return nil, errors.Newf(
		"no recording found for request hash %s for request %s. "+
			"To record this request, set the environment variable GOLLY_RECORDING_MODE=record and run the test again.",
		requestHash,
		r.URL.String(),
	)
}

func (g *Golly) passthrough(r *http.Request) (*http.Response, error) {
	return g.Doer.Do(r)
}

func NewGollyDoer(t *testing.T, recordingName string, doer httpcli.Doer) *Golly {

	var mode GollyRecordingMode
	if RecordingMode == "replay" {
		mode = RecordingModeReplay
	} else if RecordingMode == "record" {
		mode = RecordingModeRecord
	} else if RecordingMode == "passthrough" || RecordingMode == "" {
		mode = RecordingModePassthrough
	} else {
		t.Fatalf("Invalid GOLLY_RECORDING_MODE: %s", RecordingMode)
	}

	if mode == RecordingModeRecord || mode == RecordingModeReplay {
		if RecordingDir == "" {
			t.Fatalf("GOLLY_RECORDING_DIR must be set when GOLLY_RECORDING_MODE is record or replay")
		}
		stat, err := os.Stat(RecordingDir)
		if os.IsExist(err) && !stat.IsDir() {
			t.Fatalf("GOLLY_RECORDING_DIR exists and is not a directory: %s", RecordingDir)
		} else if os.IsNotExist(err) {
			if err := os.Mkdir(RecordingDir, 0755); err != nil {
				t.Fatalf("Failed to create GOLLY_RECORDING_DIR: %s", err)
			}
		}

	}

	recordingFilePath := filepath.Join(RecordingDir, recordingName+".recording.yaml")

	recordings := readRecordings(t, recordingFilePath)
	if recordings == nil {
		return nil
	}

	g := &Golly{
		T: t,
		Credentials: []TestingCredentials{
			DotcomTestingCredentials(),
			S2TestingCredentials(),
		},
		RecordingName:                 recordingName,
		RecordingFilePath:             recordingFilePath,
		RecordingMode:                 mode,
		Doer:                          doer,
		Recordings:                    recordings.Recordings,
		shouldSaveRecordingsOnCleanup: ForceSaveRecording == "true",
		Hasher:                        DefaultRequestHasher(t),
		ResponseHeaderMatcher:         DefaultResponseHeaderMatcher,
		RequestHeaderMatcher:          DefaultRequestHeaderMatcher,
	}
	t.Cleanup(g.Cleanup)
	return g
}

func DotcomTestingCredentials() TestingCredentials {
	return TestingCredentials{
		Endpoint:              "https://sourcegraph.com",
		ProductionAccessToken: DotcomAccessToken,
		RedactedToken:         "REDACTED_d5e0f0a37c9821e856b923fe14e67a605e3f6c0a517d5a4f46a4e35943ee0f6d",
	}
}

func S2TestingCredentials() TestingCredentials {
	return TestingCredentials{
		Endpoint:              "https://sourcegraph.sourcegraph.com",
		ProductionAccessToken: S2AccessToken,
		RedactedToken:         "REDACTED_964f5256e709a8c5c151a63d8696d5c7ac81604d179405864d88ff48a9232364",
	}
}

func DefaultRequestHeaderMatcher(header string, value string) bool {
	return true
}

func DefaultResponseHeaderMatcher(header string, value string) bool {
	switch header {
	case "Content-Type":
		return true
	default:
		return false
	}
}
