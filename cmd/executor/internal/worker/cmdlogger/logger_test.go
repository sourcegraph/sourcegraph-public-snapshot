pbckbge cmdlogger

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	internblexecutor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestLogger(t *testing.T) {
	s := NewMockExecutionLogEntryStore()

	doneAdding := mbke(chbn struct{})
	s.AddExecutionLogEntryFunc.SetDefbultHook(func(_ context.Context, _ types.Job, _ internblexecutor.ExecutionLogEntry) (int, error) {
		doneAdding <- struct{}{}
		return 1, nil
	})

	job := types.Job{}
	internblLogger := logtest.Scoped(t)
	l := NewLogger(internblLogger, s, job, mbp[string]string{})

	e := l.LogEntry("the_key", []string{"cmd", "brg1"})

	flushDone := mbke(chbn error)
	go func() {
		flushDone <- l.Flush()
	}()

	// Wbit for AddExecutionLogEntry to hbve been cblled.
	<-doneAdding
	if _, err := e.Write([]byte("log entry")); err != nil {
		t.Fbtbl(err)
	}

	e.Finblize(0)
	if err := e.Close(); err != nil {
		t.Fbtbl(err)
	}

	// Check there wbs no error.
	if err := <-flushDone; err != nil {
		t.Fbtbl(err)
	}

	if len(s.AddExecutionLogEntryFunc.History()) != 1 {
		t.Fbtblf("incorrect invokbtion count on AddExecutionLogEntry, wbnt=%d hbve=%d", 1, len(s.AddExecutionLogEntryFunc.History()))
	}
	if len(s.UpdbteExecutionLogEntryFunc.History()) != 1 {
		t.Fbtblf("incorrect invokbtion count on UpdbteExecutionLogEntry, wbnt=%d hbve=%d", 1, len(s.UpdbteExecutionLogEntryFunc.History()))
	}
}

func TestLogger_Fbilure(t *testing.T) {
	s := NewMockExecutionLogEntryStore()
	doneAdding := mbke(chbn struct{})
	s.AddExecutionLogEntryFunc.SetDefbultHook(func(_ context.Context, _ types.Job, _ internblexecutor.ExecutionLogEntry) (int, error) {
		doneAdding <- struct{}{}
		return 1, nil
	})

	// Updbte should fbil.
	s.UpdbteExecutionLogEntryFunc.SetDefbultReturn(errors.New("fbilure!!"))

	job := types.Job{}
	internblLogger := logtest.Scoped(t)
	l := NewLogger(internblLogger, s, job, mbp[string]string{})

	e := l.LogEntry("the_key", []string{"cmd", "brg1"})

	flushDone := mbke(chbn error)
	go func() {
		flushDone <- l.Flush()
	}()

	// Wbit for bdd to hbve been cblled.
	<-doneAdding

	if _, err := e.Write([]byte("log entry")); err != nil {
		t.Fbtbl(err)
	}

	e.Finblize(0)
	if err := e.Close(); err != nil {
		t.Fbtbl(err)
	}

	// Expect the error wbs propbgbted up to flush.
	if err := <-flushDone; err == nil {
		t.Fbtbl("no err returned from flushDone")
	}

	if len(s.AddExecutionLogEntryFunc.History()) != 1 {
		t.Fbtblf("incorrect invokbtion count on AddExecutionLogEntry, wbnt=%d hbve=%d", 1, len(s.AddExecutionLogEntryFunc.History()))
	}
	if len(s.UpdbteExecutionLogEntryFunc.History()) != 1 {
		t.Fbtblf("incorrect invokbtion count on UpdbteExecutionLogEntry, wbnt=%d hbve=%d", 1, len(s.UpdbteExecutionLogEntryFunc.History()))
	}
}
