package repos

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	workerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	Url          string `json:"url"`
	Content_type string `json:"content_type"`
	Secret       string `json:"secret"`
	Insecure_ssl string `json:"insecure_ssl"`
	Token        string `json:"token"`
	Digest       string `json:"digest,omitempty"`
}

type Payload struct {
	Name   string   `json:"name"`
	ID     int      `json:"id,omitempty"`
	Config Config   `json:"config"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

type SyncWebhookWorker struct {
	syncRequestQueue *syncRequestQueue
}

func NewSyncWebhookWorker(ctx context.Context) SyncWebhookWorker {
	syncRequestQueue := syncRequestQueue{queue: make([]*syncRequest, 0)}
	worker := SyncWebhookWorker{syncRequestQueue: &syncRequestQueue}
	return worker
}

func (worker *SyncWebhookWorker) Enqueue(repo *types.Repo) error {
	syncRequest := syncRequest{
		repo:   repo,
		secret: "secret",
		token:  "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",
	}
	ok := worker.syncRequestQueue.enqueue(syncRequest)
	if !ok {
		return errors.New("error enqueuing")
	} else {
		return nil
	}
}

type WebhookCreatingWorkerOpts struct {
	NumHandlers            int
	WorkerInterval         time.Duration
	PrometheusRegisterer   prometheus.Registerer
	CleanupOldJobs         bool
	CleanupOldJobsInterval time.Duration
}

func NewWebhookCreatingWorker(
	ctx context.Context,
	dbHandle basestore.TransactableHandle,
	handler workerutil.Handler,
	opts WebhookCreatingWorkerOpts,
) (*workerutil.Worker, *dbworker.Resetter) {
	if opts.NumHandlers == 0 {
		opts.NumHandlers = 3
	}
	if opts.WorkerInterval == 0 {
		opts.WorkerInterval = 10 * time.Second
	}
	if opts.CleanupOldJobsInterval == 0 {
		opts.CleanupOldJobsInterval = time.Hour
	}

	createWebhookJobColumns := []*sqlf.Query{
		sqlf.Sprintf("id"),
		sqlf.Sprintf("state"),
		sqlf.Sprintf("failure_message"),
		sqlf.Sprintf("started_at"),
		sqlf.Sprintf("finished_at"),
		sqlf.Sprintf("process_after"),
		sqlf.Sprintf("num_resets"),
		sqlf.Sprintf("num_failures"),
		sqlf.Sprintf("execution_logs"),
		sqlf.Sprintf("repo_id"),
		sqlf.Sprintf("repo_name"),
		sqlf.Sprintf("queued_at"),
	}

	store := workerstore.New(dbHandle, workerstore.Options{
		Name:      "webhook_creation_worker_store",
		TableName: "create_webhook_jobs",
		// ViewName:          "create_webhook_jobs_with_next_in_queue",
		Scan:              scanWebhookCreationJob,
		OrderByExpression: sqlf.Sprintf("create_webhook_jobs.queued_at"),
		ColumnExpressions: createWebhookJobColumns,
		StalledMaxAge:     30 * time.Second,
		MaxNumResets:      5,
		MaxNumRetries:     0,
	})

	worker := dbworker.NewWorker(ctx, store, handler, workerutil.WorkerOptions{
		Name:              "create_webhook_worker",
		NumHandlers:       opts.NumHandlers,
		Interval:          opts.WorkerInterval,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           newWorkerMetrics(opts.PrometheusRegisterer), // move to central pacckage
	})

	resetter := dbworker.NewResetter(store, dbworker.ResetterOptions{
		Name:     "create_webhooks_resetter",
		Interval: 5 * time.Minute,
		Metrics:  newResetterMetrics(opts.PrometheusRegisterer), // move to central package
	})

	if opts.CleanupOldJobs {
		go runJobCleaner(ctx, dbHandle, opts.CleanupOldJobsInterval) // move to central package
	}

	return worker, resetter
}

func scanWebhookCreationJob(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	fmt.Println("Scanning...")
	if err != nil {
		fmt.Println("error HERE")
		return nil, false, err
	}

	jobs, err := scanCreateJobs(rows)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}

	return &jobs[0], true, nil
}

type CreateWebhookJob struct {
	ID             int
	State          string
	FailureMessage sql.NullString
	StartedAt      sql.NullTime
	FinishedAt     sql.NullTime
	ProcessAfter   sql.NullTime
	NumResets      int
	NumFailures    int
	RepoID         int64
	RepoName       string
	QueuedAt       sql.NullTime
}

func (cw *CreateWebhookJob) RecordID() int {
	return cw.ID
}

func CreateSyncWebhook(repoURL string, secret string, token string) error { // will need secret, token, client
	fmt.Println("Creating webhook:", repoURL)

	// HOW TO GENERATE THE SECRET
	// EXTRACTING THE TOKEN FROM THE USER

	// u := "https://api.github.com/repos/susantoscott/Task-Tracker/hooks"
	parts := strings.Split(repoURL, "/")
	serviceID := parts[0]
	owner := parts[1]
	repoName := parts[2]
	url := fmt.Sprintf("https://api.%s/repos/%s/%s/hooks", serviceID, owner, repoName)
	// fmt.Println("Url:", url)
	payload := Payload{
		Name:   "web",
		Active: true,
		Config: Config{
			Url:          "https://test01/webhooks", // the url will be to /enqueue-repo-update?
			Content_type: "json",
			Secret:       secret,
			Insecure_ssl: "0",
			Token:        token,
			Digest:       "",
		},
		Events: []string{
			"push",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// fmt.Println("RespBody:", string(respBody))

	if resp.StatusCode >= 300 {
		// fmt.Println("STATUS CODE:", resp.StatusCode)
		return errors.Newf("non-200 status code, %s", err)
	}

	var obj Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		return err
	}

	return nil
}

func ListWebhooks(reponame string) []Payload {
	// fmt.Println("Listing webhooks...")

	// url := "https://api.github.com/repos/susantoscott/Task-Tracker/hooks"
	url := fmt.Sprintf("https://api.github.com/repos/susantoscott/%s/hooks", reponame)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("making new request error:", err)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	// req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Add("Authorization", "token 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")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("client do error:", err)
	}
	// fmt.Println("Status Code:", resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("readall error:", err)
	}

	var obj []Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		fmt.Println("unmarshal error:", err)
	}

	return obj
}

func DeleteWebhook(reponame string, hookID int) {
	url := fmt.Sprintf("https://api.github.com/repos/susantoscott/%s/hooks/%d", reponame, hookID)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("making new request error:", err)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", "token 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")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("client do error:", err)
	}
	fmt.Println("Status Code:", resp.StatusCode)
}

type syncRequest struct {
	repo   *types.Repo
	secret string
	token  string
}

type syncRequestQueue struct {
	mu            sync.Mutex
	queue         []*syncRequest
	notifyEnqueue chan struct{}
}

func (sq *syncRequestQueue) enqueue(syncReq syncRequest) bool {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	sq.queue = append(sq.queue, &syncReq)
	return true
}

func (sq *syncRequestQueue) dequeue() (syncRequest, bool) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	syncReq := sq.queue[0]
	sq.queue = sq.queue[1:]
	return *syncReq, true
}

func (sq *syncRequestQueue) len() int {
	return len(sq.queue)
}

// func (worker *SyncWebhookWorker) processQueue() error {
// 	sq := worker.syncRequestQueue
// 	for sq.len() > 0 {
// 		syncReq, ok := sq.dequeue()
// 		if !ok {
// 			return errors.New("Error polling!")
// 		}

// 		w, err := CreateSyncWebhook(
// 			string(syncReq.repo.Name),
// 			syncReq.secret,
// 			syncReq.token)
// 		if err != nil {
// 			return errors.New("Error creating sync webhook")
// 		}
// 		webhook := w.(Payload)
// 		fmt.Println("Have:")
// 		fmt.Printf("%+v\n", webhook)
// 	}
// 	return nil
// }
