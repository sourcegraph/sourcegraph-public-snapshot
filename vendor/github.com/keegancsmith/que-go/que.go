package que

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"
)

// Job is a single unit of work for Que to perform.
type Job struct {
	// ID is the unique database ID of the Job. It is ignored on job creation.
	ID int64

	// Queue is the name of the queue. It defaults to the empty queue "".
	Queue string

	// Priority is the priority of the Job. The default priority is 100, and a
	// lower number means a higher priority. A priority of 5 would be very
	// important.
	Priority int16

	// RunAt is the time that this job should be executed. It defaults to now(),
	// meaning the job will execute immediately. Set it to a value in the future
	// to delay a job's execution.
	RunAt time.Time

	// Type corresponds to the Ruby job_class. If you are interoperating with
	// Ruby, you should pick suitable Ruby class names (such as MyJob).
	Type string

	// Args must be the bytes of a valid JSON string
	Args []byte

	// ErrorCount is the number of times this job has attempted to run, but
	// failed with an error. It is ignored on job creation.
	ErrorCount int32

	// LastError is the error message or stack trace from the last time the job
	// failed. It is ignored on job creation.
	LastError sql.NullString

	mu    sync.Mutex
	state state
	db    *sql.DB
	tx    *sql.Tx
}

type state int

const (
	stateRunning state = iota
	stateDeleted
	stateError
)

// Delete marks this job as complete by deleting it form the database.
func (j *Job) Delete() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.state != stateRunning {
		return fmt.Errorf("job has already finished. state=%d", j.state)
	}

	_, err := j.tx.Exec(sqlDeleteJob, j.Queue, j.Priority, j.RunAt, j.ID)
	if err != nil {
		return err
	}

	return j.done(stateDeleted)
}

// done commits the transaction which releases the Postgres advisory lock on
// the job.
func (j *Job) done(finalState state) error {
	if j.state != stateRunning {
		// already marked as done
		return fmt.Errorf("que internal error. done already called on job: old=%d new=%d", j.state, finalState)
	}
	tx := j.tx
	j.tx = nil
	j.db = nil
	j.state = finalState
	return tx.Commit()
}

// Error marks the job as failed and schedules it to be reworked. An error
// message or backtrace can be provided as msg, which will be saved on the job.
// It will also increase the error count.
func (j *Job) Error(msg string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	errorCount := j.ErrorCount + 1
	delay := intPow(int(errorCount), 4) + 3 // TODO: configurable delay

	if j.state != stateRunning {
		return fmt.Errorf("job has already finished. state=%d", j.state)
	}

	_, err := j.tx.Exec(sqlSetError, errorCount, delay, msg, j.Queue, j.Priority, j.RunAt, j.ID)
	if err != nil {
		return err
	}
	return j.done(stateError)
}

// Client is a Que client that can add jobs to the queue and remove jobs from
// the queue.
type Client struct {
	db *sql.DB

	// TODO: add a way to specify default queueing options
}

// NewClient creates a new Client that uses db.
func NewClient(db *sql.DB) *Client {
	return &Client{db: db}
}

// ErrMissingType is returned when you attempt to enqueue a job with no Type
// specified.
var ErrMissingType = errors.New("job type must be specified")

// Enqueue adds a job to the queue.
func (c *Client) Enqueue(j *Job) error {
	return execEnqueue(j, c.db)
}

// EnqueueInTx adds a job to the queue within the scope of the transaction tx.
// This allows you to guarantee that an enqueued job will either be committed or
// rolled back atomically with other changes in the course of this transaction.
//
// It is the caller's responsibility to Commit or Rollback the transaction after
// this function is called.
func (c *Client) EnqueueInTx(j *Job, tx *sql.Tx) error {
	return execEnqueue(j, tx)
}

func execEnqueue(j *Job, q queryable) error {
	if j.Type == "" {
		return ErrMissingType
	}

	queue := sql.NullString{
		String: j.Queue,
		Valid:  j.Queue != "",
	}
	priority := sql.NullInt64{
		Int64: int64(j.Priority),
		Valid: j.Priority != 0,
	}
	runAt := pq.NullTime{
		Time:  j.RunAt,
		Valid: !j.RunAt.IsZero(),
	}
	args := string(j.Args)
	if args == "" {
		args = "[]"
	}

	_, err := q.Exec(sqlInsertJob, queue, priority, runAt, j.Type, args)
	return err
}

type queryable interface {
	Exec(sql string, arguments ...interface{}) (sql.Result, error)
	Query(sql string, args ...interface{}) (*sql.Rows, error)
	QueryRow(sql string, args ...interface{}) *sql.Row
}

// Maximum number of loop iterations in LockJob before giving up.  This is to
// avoid looping forever in case something is wrong.
const maxLockJobAttempts = 10

// Returned by LockJob if a job could not be retrieved from the queue after
// several attempts because of concurrently running transactions.  This error
// should not be returned unless the queue is under extremely heavy
// concurrency.
var ErrAgain = errors.New("maximum number of LockJob attempts reached")

// TODO: consider an alternate Enqueue func that also returns the newly
// enqueued Job struct. The query sqlInsertJobAndReturn was already written for
// this.

// LockJob attempts to retrieve a Job from the database in the specified queue.
// If a job is found, a session-level Postgres advisory lock is created for the
// Job's ID. If no job is found, nil will be returned instead of an error.
//
// Because Que uses session-level advisory locks, we have to hold the
// same connection throughout the process of getting a job, working it,
// deleting it, and removing the lock.
//
// After the Job has been worked, you must call either Delete() or Error() on
// it in order to indicate success or failure respectively.
func (c *Client) LockJob(queue string) (*Job, error) {
	for i := 0; i < maxLockJobAttempts; i++ {
		tx, err := c.db.Begin()
		if err != nil {
			return nil, err
		}
		j, err := c.attemptLockJob(queue, tx)
		if j == nil || err != nil {
			// ensure we rollback the transaction if we are not
			// returning a job
			err2 := tx.Rollback()
			if err2 != nil {
				return nil, err2
			}
		}
		if err != nil {
			if err == ErrAgain {
				continue
			}
			return nil, err
		}
		return j, nil
	}
	return nil, ErrAgain
}

func (c *Client) attemptLockJob(queue string, tx *sql.Tx) (*Job, error) {
	j := Job{db: c.db, tx: tx}
	err := tx.QueryRow(sqlLockJob, queue).Scan(
		&j.Queue,
		&j.Priority,
		&j.RunAt,
		&j.ID,
		&j.Type,
		&j.Args,
		&j.ErrorCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Deal with race condition. Explanation from the Ruby Que gem:
	//
	// Edge case: It's possible for the lock_job query to have
	// grabbed a job that's already been worked, if it took its MVCC
	// snapshot while the job was processing, but didn't attempt the
	// advisory lock until it was finished. Since we have the lock, a
	// previous worker would have deleted it by now, so we just
	// double check that it still exists before working it.
	//
	// Note that there is currently no spec for this behavior, since
	// I'm not sure how to reliably commit a transaction that deletes
	// the job in a separate thread between lock_job and check_job.
	var ok bool
	err = tx.QueryRow(sqlCheckJob, j.Queue, j.Priority, j.RunAt, j.ID).Scan(&ok)
	if err != nil {
		if err == sql.ErrNoRows {
			// Encountered job race condition; start over from the beginning.
			return nil, ErrAgain
		}
		return nil, err
	}
	return &j, nil
}
