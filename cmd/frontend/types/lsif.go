package types

import "time"

type LSIFDump struct {
	ID           int32     `json:"id"`
	Repository   string    `json:"repository"`
	Commit       string    `json:"commit"`
	Root         string    `json:"root"`
	VisibleAtTip bool      `json:"visibleAtTip"`
	UploadedAt   time.Time `json:"uploadedAt"`
}
