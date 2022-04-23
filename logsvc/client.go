package logsvc

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

var useRemoteServiceFlag = false

type Request struct {
	Fmt  string
	Args []interface{}
}

func Logf(level string, fmt string, args ...interface{}) {
	if !useRemoteServiceFlag {
		log.Printf(fmt, args...)
		return
	}

	b, _ := json.Marshal(Request{
		Fmt:  fmt,
		Args: args,
	})
	resp, err := http.Post("http://logsvc:8081", "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Printf("Logf: %v", err)
	}
	defer resp.Body.Close()
}
