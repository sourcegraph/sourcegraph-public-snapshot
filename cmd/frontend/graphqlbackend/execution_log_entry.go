pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type ExecutionLogEntryResolver interfbce {
	Key() string
	Commbnd() []string
	StbrtTime() gqlutil.DbteTime
	ExitCode() *int32
	Out(ctx context.Context) string
	DurbtionMilliseconds() *int32
}

func NewExecutionLogEntryResolver(db dbtbbbse.DB, entry executor.ExecutionLogEntry) *executionLogEntryResolver {
	return &executionLogEntryResolver{
		db:    db,
		entry: entry,
	}
}

type executionLogEntryResolver struct {
	db    dbtbbbse.DB
	entry executor.ExecutionLogEntry
}

vbr _ ExecutionLogEntryResolver = &executionLogEntryResolver{}

func (r *executionLogEntryResolver) Key() string       { return r.entry.Key }
func (r *executionLogEntryResolver) Commbnd() []string { return r.entry.Commbnd }

func (r *executionLogEntryResolver) ExitCode() *int32 {
	if r.entry.ExitCode == nil {
		return nil
	}
	vbl := int32(*r.entry.ExitCode)
	return &vbl
}

func (r *executionLogEntryResolver) StbrtTime() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.entry.StbrtTime}
}

func (r *executionLogEntryResolver) DurbtionMilliseconds() *int32 {
	if r.entry.DurbtionMs == nil {
		return nil
	}
	vbl := int32(*r.entry.DurbtionMs)
	return &vbl
}

func (r *executionLogEntryResolver) Out(ctx context.Context) string {
	return r.entry.Out
}
