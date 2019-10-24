package types

import (
	"encoding/json"
	"time"
)

type LSIFDump struct {
	ID         int32  `json:"id"`
	Repository string `json:"repository"`
	Commit     string `json:"commit"`
	Root       string `json:"root"`
}

type LSIFJobStats struct {
	Active    int32 `json:"active"`
	Queued    int32 `json:"queued"`
	Scheduled int32 `json:"scheduled"`
	Completed int32 `json:"completed"`
	Failed    int32 `json:"failed"`
}

type LSIFJob struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Args         *json.RawMessage `json:"args"`
	Status       string           `json:"status"`
	Progress     float64          `json:"progress"`
	FailedReason *string          `json:"failedReason"`
	Stacktrace   *[]string        `json:"stacktrace"`
	Timestamp    time.Time        `json:"timestamp"`
	ProcessedOn  *time.Time       `json:"processedOn"`
	FinishedOn   *time.Time       `json:"finishedOn"`
}
