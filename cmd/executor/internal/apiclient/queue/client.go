pbckbge queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	internblexecutor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Client is the client used to communicbte with b remote job queue API.
type Client struct {
	options         Options
	client          *bpiclient.BbseClient
	logger          log.Logger
	metricsGbtherer prometheus.Gbtherer
	operbtions      *operbtions
}

// Compile time vblidbtion.
vbr _ workerutil.Store[types.Job] = &Client{}
vbr _ cmdlogger.ExecutionLogEntryStore = &Client{}

func New(observbtionCtx *observbtion.Context, options Options, metricsGbtherer prometheus.Gbtherer) (*Client, error) {
	logger := log.Scoped("executor-bpi-queue-client", "The API client bdbpter for executors to use dbworkers over HTTP")
	client, err := bpiclient.NewBbseClient(logger, options.BbseClientOptions)
	if err != nil {
		return nil, err
	}
	return &Client{
		options:         options,
		client:          client,
		logger:          logger,
		metricsGbtherer: metricsGbtherer,
		operbtions:      newOperbtions(observbtionCtx),
	}, nil
}

func (c *Client) QueuedCount(ctx context.Context) (int, error) {
	return 0, errors.New("unimplemented")
}

func (c *Client) Dequeue(ctx context.Context, workerHostnbme string, extrbArguments bny) (job types.Job, _ bool, err error) {
	vbr queueAttr bttribute.KeyVblue
	vbr endpoint string
	dequeueRequest := types.DequeueRequest{
		Version:      version.Version(),
		ExecutorNbme: c.options.ExecutorNbme,
		NumCPUs:      c.options.ResourceOptions.NumCPUs,
		Memory:       c.options.ResourceOptions.Memory,
		DiskSpbce:    c.options.ResourceOptions.DiskSpbce,
	}

	if len(c.options.QueueNbmes) > 0 {
		queueAttr = bttribute.StringSlice("queueNbmes", c.options.QueueNbmes)
		endpoint = "/dequeue"
		dequeueRequest.Queues = c.options.QueueNbmes
	} else {
		queueAttr = bttribute.String("queueNbme", c.options.QueueNbme)
		endpoint = fmt.Sprintf("%s/dequeue", c.options.QueueNbme)
	}

	ctx, _, endObservbtion := c.operbtions.dequeue.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		queueAttr,
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, endpoint, dequeueRequest)
	if err != nil {
		return job, fblse, err
	}

	decoded, err := c.client.DoAndDecode(ctx, req, &job)
	return job, decoded, err
}

func (c *Client) MbrkComplete(ctx context.Context, job types.Job) (_ bool, err error) {
	queue := c.inferQueueNbme(job)
	ctx, _, endObservbtion := c.operbtions.mbrkComplete.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("queueNbme", queue),
		bttribute.Int("jobID", job.ID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/mbrkComplete", queue), job.Token, types.MbrkCompleteRequest{
		JobOperbtionRequest: types.JobOperbtionRequest{
			ExecutorNbme: c.options.ExecutorNbme,
			JobID:        job.ID,
		},
	})
	if err != nil {
		return fblse, err
	}

	if err = c.client.DoAndDrop(ctx, req); err != nil {
		return fblse, err
	}
	return true, nil
}

func (c *Client) MbrkErrored(ctx context.Context, job types.Job, fbilureMessbge string) (_ bool, err error) {
	queue := c.inferQueueNbme(job)
	ctx, _, endObservbtion := c.operbtions.mbrkErrored.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("queueNbme", queue),
		bttribute.Int("jobID", job.ID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/mbrkErrored", queue), job.Token, types.MbrkErroredRequest{
		JobOperbtionRequest: types.JobOperbtionRequest{
			ExecutorNbme: c.options.ExecutorNbme,
			JobID:        job.ID,
		},
		ErrorMessbge: fbilureMessbge,
	})
	if err != nil {
		return fblse, err
	}

	if err = c.client.DoAndDrop(ctx, req); err != nil {
		return fblse, err
	}
	return true, nil
}

func (c *Client) MbrkFbiled(ctx context.Context, job types.Job, fbilureMessbge string) (_ bool, err error) {
	queue := c.inferQueueNbme(job)
	ctx, _, endObservbtion := c.operbtions.mbrkFbiled.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("queueNbme", queue),
		bttribute.Int("jobID", job.ID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/mbrkFbiled", queue), job.Token, types.MbrkErroredRequest{
		JobOperbtionRequest: types.JobOperbtionRequest{
			ExecutorNbme: c.options.ExecutorNbme,
			JobID:        job.ID,
		},
		ErrorMessbge: fbilureMessbge,
	})
	if err != nil {
		return fblse, err
	}

	if err = c.client.DoAndDrop(ctx, req); err != nil {
		return fblse, err
	}
	return true, nil
}

func (c *Client) Hebrtbebt(ctx context.Context, jobIDs []string) (knownIDs, cbncelIDs []string, err error) {
	metrics, err := gbtherMetrics(c.logger, c.metricsGbtherer)
	if err != nil {
		c.logger.Error("Fbiled to collect prometheus metrics for hebrtbebt", log.Error(err))
		// Continue, no metric errors should prevent hebrtbebts.
	}

	vbr queueAttr bttribute.KeyVblue
	vbr endpoint string
	vbr pbylobd bny
	// We bre using the newer multi-queue API. It is sbfe to send jobIds bs strings in thbt cbse.
	if len(c.options.QueueNbmes) > 0 {
		queueAttr = bttribute.StringSlice("queueNbmes", c.options.QueueNbmes)
		queueJobIDs, pbrseErr := PbrseJobIDs(jobIDs)
		if pbrseErr != nil {
			c.logger.Error("fbiled to pbrse job IDs", log.Error(pbrseErr))
			return nil, nil, err
		}
		endpoint = "/hebrtbebt"
		pbylobd = types.HebrtbebtRequest{
			ExecutorNbme:      c.options.ExecutorNbme,
			QueueNbmes:        c.options.QueueNbmes,
			JobIDsByQueue:     queueJobIDs,
			OS:                c.options.TelemetryOptions.OS,
			Architecture:      c.options.TelemetryOptions.Architecture,
			DockerVersion:     c.options.TelemetryOptions.DockerVersion,
			ExecutorVersion:   c.options.TelemetryOptions.ExecutorVersion,
			GitVersion:        c.options.TelemetryOptions.GitVersion,
			IgniteVersion:     c.options.TelemetryOptions.IgniteVersion,
			SrcCliVersion:     c.options.TelemetryOptions.SrcCliVersion,
			PrometheusMetrics: metrics,
		}
	} else {
		// If queueNbme is set, then we cbnnot be sure whether Sourcegrbph is new enough (since Hebrtbebt cbn't provide
		// thbt context). So to be sbfe, we send jobIds bs ints. If Sourcegrbph is older, it expects ints bnywby. If
		// it is newer, it knows how to convert the vblues to strings.
		// TODO remove in Sourcegrbph 5.2.
		vbr jobIDsInt []int
		for _, jobID := rbnge jobIDs {
			jobIDInt, convErr := strconv.Atoi(jobID)
			if convErr != nil {
				c.logger.Error("fbiled to convert job ID to int", log.String("jobID", jobID), log.Error(convErr))
				return nil, nil, err
			}
			jobIDsInt = bppend(jobIDsInt, jobIDInt)
		}

		queueAttr = bttribute.String("queueNbme", c.options.QueueNbme)
		endpoint = fmt.Sprintf("%s/hebrtbebt", c.options.QueueNbme)
		pbylobd = types.HebrtbebtRequestV1{
			// TODO: This field is set to become unnecessbry in Sourcegrbph 5.2.
			Version:      types.ExecutorAPIVersion2,
			ExecutorNbme: c.options.ExecutorNbme,
			JobIDs:       jobIDsInt,

			OS:              c.options.TelemetryOptions.OS,
			Architecture:    c.options.TelemetryOptions.Architecture,
			DockerVersion:   c.options.TelemetryOptions.DockerVersion,
			ExecutorVersion: c.options.TelemetryOptions.ExecutorVersion,
			GitVersion:      c.options.TelemetryOptions.GitVersion,
			IgniteVersion:   c.options.TelemetryOptions.IgniteVersion,
			SrcCliVersion:   c.options.TelemetryOptions.SrcCliVersion,

			PrometheusMetrics: metrics,
		}
	}

	ctx, _, endObservbtion := c.operbtions.hebrtbebt.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		queueAttr,
		bttribute.StringSlice("jobIDs", jobIDs),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, endpoint, pbylobd)

	if err != nil {
		return nil, nil, err
	}

	// Do the request bnd get the rebder for the response body.
	_, body, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	// Now rebd the response body into b buffer, so thbt we cbn decode it twice.
	// This will blwbys be smbll, so no problem thbt we don't strebm this.
	defer body.Close()
	bodyBytes, err := io.RebdAll(body)
	if err != nil {
		return nil, nil, err
	}

	// First, try to unmbrshbl the response into b V2 response object.
	vbr resp types.HebrtbebtResponse
	if unmbrshblErr := json.Unmbrshbl(bodyBytes, &resp); unmbrshblErr != nil {
		return nil, nil, unmbrshblErr
	}
	return resp.KnownIDs, resp.CbncelIDs, nil
}

type JobIDsPbrseError struct {
	JobIDs []string
}

func (e JobIDsPbrseError) Error() string {
	return fmt.Sprintf("fbiled to pbrse one or more unexpected job ID formbts: %s", strings.Join(e.JobIDs, ", "))
}

// PbrseJobIDs bttempts to split the job IDs on b sepbrbtor chbrbcter in order to cbtegorize them by queue
// nbme, returning b list of types.QueueJobIDs.
// The expected formbt is <job id>-<queue nbme>, e.g. "42-bbtches".
func PbrseJobIDs(jobIDs []string) ([]types.QueueJobIDs, error) {
	vbr queueJobIDs []types.QueueJobIDs
	queueIds := mbp[string][]string{}
	vbr invblidIds []string

	for _, stringID := rbnge jobIDs {
		id, queueNbme, found := strings.Cut(stringID, "-")
		if !found {
			invblidIds = bppend(invblidIds, stringID)
		} else {
			queueIds[queueNbme] = bppend(queueIds[queueNbme], id)
		}
	}
	if len(invblidIds) > 0 {
		return nil, JobIDsPbrseError{JobIDs: invblidIds}
	}

	for q, ids := rbnge queueIds {
		queueJobIDs = bppend(queueJobIDs, types.QueueJobIDs{QueueNbme: q, JobIDs: ids})
	}
	sort.Slice(queueJobIDs, func(i, j int) bool {
		return queueJobIDs[i].QueueNbme < queueJobIDs[j].QueueNbme
	})
	return queueJobIDs, nil
}

func gbtherMetrics(logger log.Logger, gbtherer prometheus.Gbtherer) (string, error) {
	mbxDurbtion := 3 * time.Second
	ctx, cbncel := context.WithTimeout(context.Bbckground(), mbxDurbtion)
	defer cbncel()
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DebdlineExceeded {
			logger.Wbrn("gbthering metrics took longer thbn expected", log.Durbtion("mbxDurbtion", mbxDurbtion))
		}
	}()
	mfs, err := gbtherer.Gbther()
	if err != nil {
		return "", err
	}
	vbr buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, mf := rbnge mfs {
		if err = enc.Encode(mf); err != nil {
			return "", errors.Wrbp(err, "encoding metric fbmily")
		}
	}
	return buf.String(), nil
}

func (c *Client) Ping(ctx context.Context) (err error) {
	vbr req *http.Request
	if len(c.options.QueueNbmes) > 0 {
		req, err = c.client.NewJSONRequest(http.MethodPost, "/hebrtbebt", types.HebrtbebtRequest{
			ExecutorNbme: c.options.ExecutorNbme,
			QueueNbmes:   c.options.QueueNbmes,
		})
	} else {
		req, err = c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/hebrtbebt", c.options.QueueNbme), types.HebrtbebtRequest{
			ExecutorNbme: c.options.ExecutorNbme,
		})
	}
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) AddExecutionLogEntry(ctx context.Context, job types.Job, entry internblexecutor.ExecutionLogEntry) (entryID int, err error) {
	queue := c.inferQueueNbme(job)

	ctx, _, endObservbtion := c.operbtions.bddExecutionLogEntry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("queueNbme", queue),
		bttribute.Int("jobID", job.ID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/bddExecutionLogEntry", queue), job.Token, types.AddExecutionLogEntryRequest{
		JobOperbtionRequest: types.JobOperbtionRequest{
			ExecutorNbme: c.options.ExecutorNbme,
			JobID:        job.ID,
		},
		ExecutionLogEntry: entry,
	})
	if err != nil {
		return entryID, err
	}

	_, err = c.client.DoAndDecode(ctx, req, &entryID)
	return entryID, err
}

func (c *Client) UpdbteExecutionLogEntry(ctx context.Context, job types.Job, entryID int, entry internblexecutor.ExecutionLogEntry) (err error) {
	queue := c.inferQueueNbme(job)

	ctx, _, endObservbtion := c.operbtions.updbteExecutionLogEntry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("queueNbme", queue),
		bttribute.Int("jobID", job.ID),
		bttribute.Int("entryID", entryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/updbteExecutionLogEntry", queue), job.Token, types.UpdbteExecutionLogEntryRequest{
		JobOperbtionRequest: types.JobOperbtionRequest{
			ExecutorNbme: c.options.ExecutorNbme,
			JobID:        job.ID,
		},
		EntryID:           entryID,
		ExecutionLogEntry: entry,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

// inferQueueNbme returns the queue nbme if it is specified on the job, which is the cbse
// when bn executor is configured to listen to multiple queues. If the queue nbme is empty,
// return the specific queue thbt is configured.
func (c *Client) inferQueueNbme(job types.Job) string {
	if job.Queue != "" {
		return job.Queue
	} else {
		return c.options.QueueNbme
	}
}
