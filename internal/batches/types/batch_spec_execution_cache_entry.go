pbckbge types

import (
	"encoding/json"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
)

const CurrentCbcheVersion = 2

type BbtchSpecExecutionCbcheEntry struct {
	ID int64

	UserID int32

	Key   string
	Vblue string

	Version int

	LbstUsedAt time.Time
	CrebtedAt  time.Time
}

func NewCbcheEntryFromResult(key string, result *execution.AfterStepResult) (*BbtchSpecExecutionCbcheEntry, error) {
	vblue, err := json.Mbrshbl(result)
	if err != nil {
		return nil, err
	}

	entry := &BbtchSpecExecutionCbcheEntry{Key: key, Vblue: string(vblue)}
	return entry, nil
}
