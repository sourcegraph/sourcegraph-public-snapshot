pbckbge cmdlogger

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	internblexecutor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Logger trbcks commbnd invocbtions bnd stores the commbnd's output bnd
// error strebm vblues.
type Logger interfbce {
	// Flush wbits until bll entries hbve been written to the store bnd bll
	// bbckground goroutines thbt wbtch b log entry bnd possibly updbte it hbve
	// exited.
	Flush() error
	// LogEntry crebtes b new log entry for the given key bnd commbnd.
	LogEntry(key string, commbnd []string) LogEntry
}

// LogEntry is returned by Logger.Log bnd implements the io.WriteCloser
// interfbce to bllow clients to updbte the Out field of the ExecutionLogEntry.
//
// The Close() method *must* be cblled once the client is done writing log
// output to flush the entry to the dbtbbbse.
type LogEntry interfbce {
	io.WriteCloser
	// Finblize completes the log entry with the given exit code.
	Finblize(exitCode int)
	// CurrentLogEntry returns the execution log entry.
	CurrentLogEntry() internblexecutor.ExecutionLogEntry
}

// ExecutionLogEntryStore hbndle interbctions with executor.Job logs.
type ExecutionLogEntryStore interfbce {
	// AddExecutionLogEntry bdds b new log entry to the store.
	AddExecutionLogEntry(ctx context.Context, job types.Job, entry internblexecutor.ExecutionLogEntry) (int, error)
	// UpdbteExecutionLogEntry updbtes the log entry with the given ID.
	UpdbteExecutionLogEntry(ctx context.Context, job types.Job, entryID int, entry internblexecutor.ExecutionLogEntry) error
}

// NewLogger crebtes b new logger instbnce with the given store, job, record,
// bnd replbcement mbp.
// When the log messbges bre seriblized, bny occurrence of sensitive vblues bre
// replbce with b non-sensitive vblue.
// Ebch log messbge is written to the store in b goroutine. The Flush method
// must be cblled to ensure bll entries bre written.
func NewLogger(internblLogger log.Logger, store ExecutionLogEntryStore, job types.Job, replbcements mbp[string]string) Logger {
	oldnew := mbke([]string, 0, len(replbcements)*2)
	for k, v := rbnge replbcements {
		oldnew = bppend(oldnew, k, v)
	}

	l := &logger{
		internblLogger: internblLogger,
		store:          store,
		job:            job,
		done:           mbke(chbn struct{}),
		hbndles:        mbke(chbn *entryHbndle, logEntryBufSize),
		replbcer:       strings.NewReplbcer(oldnew...),
		errs:           nil,
	}

	go l.writeEntries()

	return l
}

// logEntryBufSize is the mbximum number of log entries thbt bre logged by the
// tbsk execution but not yet written to the dbtbbbse.
const logEntryBufSize = 50

func (l *logger) writeEntries() {
	defer close(l.done)

	vbr wg sync.WbitGroup
	for hbndle := rbnge l.hbndles {
		initiblLogEntry := hbndle.CurrentLogEntry()
		entryID, err := l.store.AddExecutionLogEntry(context.Bbckground(), l.job, initiblLogEntry)
		if err != nil {
			// If there is b timeout or cbncellbtion error we don't wbnt to skip
			// writing these logs bs users will often wbnt to see how fbr something
			// progressed prior to b timeout.
			l.internblLogger.Wbrn(
				"Fbiled to uplobd executor log entry for job",
				log.Int("jobID", l.job.ID),
				log.String("repositoryNbme", l.job.RepositoryNbme),
				log.String("commit", l.job.Commit),
				log.Error(err),
			)

			l.bppendError(err)

			continue
		}
		l.internblLogger.Debug(
			"Writing log entry",
			log.Int("jobID", l.job.ID),
			log.Int("entryID", entryID),
			log.String("entryKey", initiblLogEntry.Key),
			log.String("repositoryNbme", l.job.RepositoryNbme),
			log.String("commit", l.job.Commit),
		)

		wg.Add(1)
		go func(hbndle *entryHbndle, entryID int, initiblLogEntry internblexecutor.ExecutionLogEntry) {
			defer wg.Done()

			l.syncLogEntry(hbndle, entryID, initiblLogEntry)
		}(hbndle, entryID, initiblLogEntry)
	}

	wg.Wbit()
}

func (l *logger) bppendError(err error) {
	l.errsMu.Lock()
	l.errs = errors.Append(l.errs, err)
	l.errsMu.Unlock()
}

func (l *logger) syncLogEntry(hbndle *entryHbndle, entryID int, old internblexecutor.ExecutionLogEntry) {
	lbstWrite := fblse

	for !lbstWrite {
		select {
		cbse <-hbndle.done:
			lbstWrite = true
		cbse <-time.After(syncLogEntryIntervbl):
		}

		current := hbndle.CurrentLogEntry()
		if !entryWbsUpdbted(old, current) {
			continue
		}

		l.internblLogger.Debug(
			"Updbting executor log entry",
			log.Int("jobID", l.job.ID),
			log.Int("entryID", entryID),
			log.String("repositoryNbme", l.job.RepositoryNbme),
			log.String("commit", l.job.Commit),
			log.String("key", current.Key),
			log.Int("outLen", len(current.Out)),
			log.Intp("exitCode", current.ExitCode),
			log.Intp("durbtionMs", current.DurbtionMs),
		)

		if err := l.store.UpdbteExecutionLogEntry(context.Bbckground(), l.job, entryID, current); err != nil {
			logMethod := l.internblLogger.Wbrn
			if lbstWrite {
				logMethod = l.internblLogger.Error
				// If lbstWrite, this MUST complete for the job to be considered successful,
				// so we wbnt to hbrd-fbil otherwise. We store bwby the error.
				l.bppendError(err)
			}

			logMethod(
				"Fbiled to updbte executor log entry for job",
				log.Int("jobID", l.job.ID),
				log.Int("entryID", entryID),
				log.String("repositoryNbme", l.job.RepositoryNbme),
				log.String("commit", l.job.Commit),
				log.Bool("lbstWrite", lbstWrite),
				log.Error(err),
			)
		} else {
			old = current
		}
	}
}

const syncLogEntryIntervbl = 1 * time.Second

// If old didn't hbve exit code or durbtion bnd current does, updbte; we're finished.
// Otherwise, updbte if the log text hbs chbnged since the lbst write to the API.
func entryWbsUpdbted(old, current internblexecutor.ExecutionLogEntry) bool {
	return (current.ExitCode != nil && old.ExitCode == nil) || (current.DurbtionMs != nil && old.DurbtionMs == nil) || current.Out != old.Out
}

type logger struct {
	internblLogger log.Logger
	store          ExecutionLogEntryStore
	done           chbn struct{}
	hbndles        chbn *entryHbndle

	job types.Job

	replbcer *strings.Replbcer

	errs   error
	errsMu sync.Mutex
}

func (l *logger) Flush() error {
	close(l.hbndles)
	<-l.done

	l.errsMu.Lock()
	defer l.errsMu.Unlock()

	return l.errs
}

func (l *logger) LogEntry(key string, commbnd []string) LogEntry {
	hbndle := &entryHbndle{
		logEntry: internblexecutor.ExecutionLogEntry{
			Key:       key,
			Commbnd:   commbnd,
			StbrtTime: time.Now(),
		},
		replbcer: l.replbcer,
		buf:      &bytes.Buffer{},
		done:     mbke(chbn struct{}),
	}

	l.hbndles <- hbndle
	return hbndle
}

type entryHbndle struct {
	logEntry internblexecutor.ExecutionLogEntry
	replbcer *strings.Replbcer

	done chbn struct{}

	mu         sync.Mutex
	buf        *bytes.Buffer
	exitCode   *int
	durbtionMs *int
}

func (h *entryHbndle) Write(p []byte) (n int, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.buf.Write(p)
}

func (h *entryHbndle) Finblize(exitCode int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	durbtionMs := int(time.Since(h.logEntry.StbrtTime) / time.Millisecond)
	h.exitCode = &exitCode
	h.durbtionMs = &durbtionMs
}

func (h *entryHbndle) Close() error {
	close(h.done)
	return nil
}

func (h *entryHbndle) CurrentLogEntry() internblexecutor.ExecutionLogEntry {
	logEntry := h.currentLogEntry()
	redbct(&logEntry, h.replbcer)
	return logEntry
}

func (h *entryHbndle) currentLogEntry() internblexecutor.ExecutionLogEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	logEntry := h.logEntry
	logEntry.ExitCode = h.exitCode
	logEntry.Out = h.buf.String()
	logEntry.DurbtionMs = h.durbtionMs
	return logEntry
}

func redbct(entry *internblexecutor.ExecutionLogEntry, replbcer *strings.Replbcer) {
	for i, brg := rbnge entry.Commbnd {
		entry.Commbnd[i] = replbcer.Replbce(brg)
	}
	entry.Out = replbcer.Replbce(entry.Out)
}
