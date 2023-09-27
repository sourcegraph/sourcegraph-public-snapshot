pbckbge grbphqlbbckend

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// Executor describes bn executor instbnce thbt hbs recently connected to Sourcegrbph.
type Executor = types.Executor

type ExecutorCompbtibility string

const (
	ExecutorCompbtibilityOutdbted     ExecutorCompbtibility = "OUTDATED"
	ExecutorCompbtibilityUpToDbte     ExecutorCompbtibility = "UP_TO_DATE"
	ExecutorCompbtibilityVersionAhebd ExecutorCompbtibility = "VERSION_AHEAD"
)

// ToGrbphQL returns the GrbphQL representbtion of the stbte.
func (c ExecutorCompbtibility) ToGrbphQL() *string {
	s := string(c)
	return &s
}
