package golly

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
)

func (r *Golly) Find(requestHash string) *yamlRecording {
	for _, rec := range r.Recordings {
		if requestHash == rec.Hash {
			rec.isActive = true
			return rec
		}
	}
	return nil
}

func (r *Golly) HashOrPanic(req *http.Request, requestBody []byte) string {
	hash, err := r.Hasher(req, requestBody)
	if err != nil {
		panic(fmt.Sprintf("Failed to hash request: %v", err))
	}
	return hash
}

func (g *Golly) AddRecording(req *http.Request, requestBody []byte, hash string, res *http.Response) *yamlRecording {
	recording := &yamlRecording{
		Hash:     hash,
		Request:  g.newYamlRequest(req, requestBody),
		Response: g.newYamlResponse(res),
		isActive: true,
	}
	g.Recordings = append(g.Recordings, recording)
	g.shouldSaveRecordingsOnCleanup = true
	return recording
}

func (g *Golly) Cleanup() {
	if g.T.Failed() {
		return
	}
	if !g.shouldSaveRecordingsOnCleanup {
		return
	}
	g.doSave()
}

type yamlRecordings struct {
	Recordings []*yamlRecording `yaml:"recordings"`
}

type yamlRecording struct {
	Hash     string        `yaml:"hash"`
	Request  *yamlRequest  `yaml:"request"`
	Response *yamlResponse `yaml:"response"`
	// isActive is true if this recording has been recorded or replayed.
	isActive bool
}

func readyBodyIntoMemory(t *testing.T, req *http.Request) []byte {
	if req.Body == nil {
		return []byte{}
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Failed to read request body: %v", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}

func (g *Golly) newYamlRequest(req *http.Request, requestBody []byte) *yamlRequest {
	body := string(requestBody)
	if contentType := req.Header.Get("Content-Type"); strings.HasPrefix(contentType, "application/json") {
		var jsonBody interface{}
		if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
			// Format as YAML for readability. We don't deserialize the
			// request body back when reading recordings so it doesn't
			// matter if we use JSON or YAML here.
			if formattedYAML, err := yaml.Marshal(jsonBody); err == nil {
				body = string(formattedYAML)
			}
		}
	}
	return &yamlRequest{
		RecordingDate: time.Now().Format(time.RFC3339),
		URL:           req.URL.String(),
		Method:        req.Method,
		Headers:       g.newYamlRequestHeaders(req),
		Body:          body,
	}
}

type yamlRequest struct {
	RecordingDate string       `yaml:"recording_date"`
	URL           string       `yaml:"url"`
	Method        string       `yaml:"method"`
	Headers       []yamlHeader `yaml:"headers"`
	Body          string       `yaml:"body"`
}

func (g *Golly) newYamlRequestHeaders(request *http.Request) []yamlHeader {
	var yamlHeaders []yamlHeader
	for key, values := range request.Header {
		for _, value := range values {
			if g.RequestHeaderMatcher(key, value) {
				h := yamlHeader{Key: key, Value: value}
				if key == "Authorization" {
					h.Value = redactAuthorizationHeader(value)
				}
				yamlHeaders = append(yamlHeaders, h)
			}
		}
	}
	return yamlHeaders
}

func (g *Golly) newYamlResponseHeaders(headers http.Header) []yamlHeader {
	var yamlHeaders []yamlHeader
	for key, values := range headers {
		for _, value := range values {
			if g.ResponseHeaderMatcher(key, value) {
				yamlHeaders = append(yamlHeaders, yamlHeader{Key: key, Value: value})
			}
		}
	}
	return yamlHeaders
}

type yamlHeader struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func (g *Golly) newYamlResponse(res *http.Response) *yamlResponse {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		g.T.Fatalf("Failed to read response body: %v", err)
	}
	res.Body = io.NopCloser(bytes.NewBuffer(body))
	return &yamlResponse{
		StatusCode: res.StatusCode,
		Body:       string(body),
		Headers:    g.newYamlResponseHeaders(res.Header),
	}
}

type yamlResponse struct {
	StatusCode int          `yaml:"status_code"`
	Body       string       `yaml:"body"`
	Headers    []yamlHeader `yaml:"headers"`
}

func (r *yamlResponse) HttpResponse() *http.Response {
	header := http.Header{}
	for _, h := range r.Headers {
		header.Add(h.Key, h.Value)
	}
	res := &http.Response{
		StatusCode: r.StatusCode,
		Header:     header,
	}
	res.Body = io.NopCloser(bytes.NewBufferString(r.Body))
	return res
}

func (r *Golly) doSave() {
	filePath := filepath.Join(r.RecordingFilePath)
	recordings := &yamlRecordings{}
	for _, rec := range r.Recordings {
		if rec.isActive {
			recordings.Recordings = append(recordings.Recordings, rec)
		}
	}
	data, err := yaml.Marshal(recordings)
	if err != nil {
		r.T.Fatalf("failed to marshal recordings to YAML: %v", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		r.T.Fatalf("failed to write recordings to file %s: %v", filePath, err)
	}
}

func readRecordings(t *testing.T, recordingFilePath string) *yamlRecordings {
	filePath := filepath.Join(recordingFilePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Empty file means no recordings
		return &yamlRecordings{}
	}

	rec := &yamlRecordings{}
	err = yaml.Unmarshal(data, rec)
	if err != nil {
		t.Fatalf("Failed to unmarshal recordings from file %s: %v", filePath, err)
		return nil
	}
	return rec
}

func redactAuthorizationHeader(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte("redacted_" + value))
	return fmt.Sprintf("token REDACTED_%x", hasher.Sum(nil))
}
