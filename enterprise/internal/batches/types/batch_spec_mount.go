package types

import "time"

type BatchSpecMount struct {
	ID          int64
	RandID      string
	BatchSpecID int64

	FileName string
	Path     string
	Size     int64
	Modified time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a BatchSpecMount.
func (b *BatchSpecMount) Clone() *BatchSpecMount {
	clone := *b
	return &clone
}
