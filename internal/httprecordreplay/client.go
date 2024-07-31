package httprecordreplay

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func EnableHTTPRecordReplay() {
	switch mode := os.Getenv("HTTP_RECORD_REPLAY_MODE"); mode {
	case "", "passthrough":
		return
	case "replay":
		setupReplay()
	case "record":
		setupRecord()
	default:
		panic("Invalid HTTP_RECORD_REPLAY_MODE: " + mode)
	}
}

func setupRecord() {
	httpcli.GlobalDoerMock = httpcli.DoerMock{
		DoFunc: func(underlying httpcli.DoerFunc, req *http.Request) (*http.Response, error) {
			fmt.Println("request:", req.URL, "headers:", req.Header)
			return underlying(req)
		},
	}
	// TODO
}

func setupReplay() {
	// TODO
}
