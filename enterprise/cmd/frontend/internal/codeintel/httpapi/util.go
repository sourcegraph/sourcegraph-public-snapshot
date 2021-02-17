package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

func sanitizeRoot(s string) string {
	if s == "" || s == "/" {
		return ""
	}
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s
}

func hasQuery(r *http.Request, name string) bool {
	return r.URL.Query().Get(name) != ""
}

func getQuery(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func getQueryInt(r *http.Request, name string) int {
	value, _ := strconv.Atoi(r.URL.Query().Get(name))
	return value
}

// copyAll writes the contents of r to w and logs on write failure.
func copyAll(w http.ResponseWriter, r io.Reader) {
	if _, err := io.Copy(w, r); err != nil {
		log15.Error("Failed to write payload to client", "error", err)
	}
}

// writeJSON writes the JSON-encoded payload to w and logs on write failure.
// If there is an encoding error, then a 500-level status is written to w.
func writeJSON(w http.ResponseWriter, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log15.Error("Failed to serialize result", "error", err)
		http.Error(w, fmt.Sprintf("failed to serialize result: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	copyAll(w, bytes.NewReader(data))
}

func formatAWSError(err awserr.Error) string {
	var s3Err s3manager.MultiUploadFailure
	if errors.As(err, &s3Err) {
		if s3Err.OrigErr() != nil {
			return s3Err.OrigErr().Error()
		}

		return s3Err.Error()
	}

	return err.Message()
}
