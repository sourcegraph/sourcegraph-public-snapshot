package main

import (
	"encoding/json"
	"os"
	"time"
)

type batchesLogEvent struct {
	Operation string `json:"operation"` // "PREPARING_DOCKER_IMAGES"

	Timestamp time.Time `json:"timestamp"`

	Status  string `json:"status"`            // "STARTED", "PROGRESS", "SUCCESS", "FAILURE"
	Message string `json:"message,omitempty"` // "70% done"
}

func logOperationStart(op, msg string) {
	logEvent(batchesLogEvent{Operation: op, Status: "STARTED", Message: msg})
}

func logOperationSuccess(op, msg string) {
	logEvent(batchesLogEvent{Operation: op, Status: "SUCCESS", Message: msg})
}

func logOperationFailure(op, msg string) {
	logEvent(batchesLogEvent{Operation: op, Status: "FAILURE", Message: msg})
}

func logOperationProgress(op, msg string) {
	logEvent(batchesLogEvent{Operation: op, Status: "PROGRESS", Message: msg})
}

func logEvent(e batchesLogEvent) {
	e.Timestamp = time.Now().UTC().Truncate(time.Millisecond)
	json.NewEncoder(os.Stdout).Encode(e)
}
