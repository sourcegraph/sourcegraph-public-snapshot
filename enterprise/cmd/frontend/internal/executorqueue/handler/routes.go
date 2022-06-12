package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/gorilla/mux"
	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/log"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SetupRoutes registers all route handlers required for all configured executor
// queues with the given router.
func SetupRoutes(executorStore executor.Store, metricsStore metricsstore.DistributedStore, queueOptionsMap []QueueOptions, router *mux.Router) {
	for _, queueOptions := range queueOptionsMap {
		h := newHandler(executorStore, metricsStore, queueOptions)

		subRouter := router.PathPrefix(fmt.Sprintf("/{queueName:(?:%s)}/", regexp.QuoteMeta(queueOptions.Name))).Subrouter()
		routes := map[string]func(w http.ResponseWriter, r *http.Request){
			"dequeue":                 h.handleDequeue,
			"addExecutionLogEntry":    h.handleAddExecutionLogEntry,
			"updateExecutionLogEntry": h.handleUpdateExecutionLogEntry,
			"markComplete":            h.handleMarkComplete,
			"markErrored":             h.handleMarkErrored,
			"markFailed":              h.handleMarkFailed,
			"heartbeat":               h.handleHeartbeat,
			"canceled":                h.handleCanceled,
		}
		for path, handler := range routes {
			subRouter.Path(fmt.Sprintf("/%s", path)).Methods("POST").HandlerFunc(handler)
		}
	}
}

// POST /{queueName}/dequeue
func (h *handler) handleDequeue(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.DequeueRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		job, dequeued, err := h.dequeue(r.Context(), payload.ExecutorName)
		if !dequeued {
			return http.StatusNoContent, nil, err
		}

		return http.StatusOK, job, err
	})
}

// POST /{queueName}/addExecutionLogEntry
func (h *handler) handleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.AddExecutionLogEntryRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		id, err := h.addExecutionLogEntry(r.Context(), payload.ExecutorName, payload.JobID, payload.ExecutionLogEntry)
		return http.StatusOK, id, err
	})
}

// POST /{queueName}/updateExecutionLogEntry
func (h *handler) handleUpdateExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.UpdateExecutionLogEntryRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.updateExecutionLogEntry(r.Context(), payload.ExecutorName, payload.JobID, payload.EntryID, payload.ExecutionLogEntry)
		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/markComplete
func (h *handler) handleMarkComplete(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.MarkCompleteRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.markComplete(r.Context(), payload.ExecutorName, payload.JobID)
		if err == ErrUnknownJob {
			return http.StatusNotFound, nil, nil
		}

		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/markErrored
func (h *handler) handleMarkErrored(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.MarkErroredRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.markErrored(r.Context(), payload.ExecutorName, payload.JobID, payload.ErrorMessage)
		if err == ErrUnknownJob {
			return http.StatusNotFound, nil, nil
		}

		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/markFailed
func (h *handler) handleMarkFailed(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.MarkErroredRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.markFailed(r.Context(), payload.ExecutorName, payload.JobID, payload.ErrorMessage)
		if err == ErrUnknownJob {
			return http.StatusNotFound, nil, nil
		}

		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/heartbeat
func (h *handler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.HeartbeatRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		executor := types.Executor{
			Hostname:        payload.ExecutorName,
			QueueName:       h.QueueOptions.Name,
			OS:              payload.OS,
			Architecture:    payload.Architecture,
			DockerVersion:   payload.DockerVersion,
			ExecutorVersion: payload.ExecutorVersion,
			GitVersion:      payload.GitVersion,
			IgniteVersion:   payload.IgniteVersion,
			SrcCliVersion:   payload.SrcCliVersion,
		}

		// Handle metrics in the background, this should not delay the heartbeat response being
		// delivered. It is critical for keeping jobs alive.
		go func() {
			metrics, err := decodeAndLabelMetrics(payload.PrometheusMetrics, payload.ExecutorName, payload.ExecutorVersion)
			if err != nil {
				// Just log the error but don't panic. The heartbeat is more important.
				h.logger.Error("failed to decode metrics and apply labels for executor heartbeat", log.Error(err))
			}

			if err := h.metricsStore.Ingest(payload.ExecutorName, metrics); err != nil {
				// Just log the error but don't panic. The heartbeat is more important.
				h.logger.Error("failed to ingest metrics for executor heartbeat", log.Error(err))
			}
		}()

		unknownIDs, err := h.heartbeat(r.Context(), executor, payload.JobIDs)
		return http.StatusOK, unknownIDs, err
	})
}

// POST /{queueName}/canceled
func (h *handler) handleCanceled(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.CanceledRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		canceledIDs, err := h.canceled(r.Context(), payload.ExecutorName)
		return http.StatusOK, canceledIDs, err
	})
}

type errorResponse struct {
	Error string `json:"error"`
}

// wrapHandler decodes the request body into the given payload pointer, then calls the given
// handler function. If the body cannot be decoded, a 400 BadRequest is returned and the handler
// function is not called. If the handler function returns an error, a 500 Internal Server Error
// is returned. Otherwise, the response status will match the status code value returned from the
// handler, and the payload value returned from the handler is encoded and written to the
// response body.
func (h *handler) wrapHandler(w http.ResponseWriter, r *http.Request, payload any, handler func() (int, any, error)) {
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal payload: %s", err.Error()), http.StatusBadRequest)
		return
	}

	status, payload, err := handler()
	if err != nil {
		log15.Error("Handler returned an error", "err", err)

		status = http.StatusInternalServerError
		payload = errorResponse{Error: err.Error()}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log15.Error("Failed to serialize payload", "err", err)
		http.Error(w, fmt.Sprintf("Failed to serialize payload: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)

	if status != http.StatusNoContent {
		_, _ = io.Copy(w, bytes.NewReader(data))
	}
}

// decodeAndLabelMetrics decodes the text serialized prometheus metrics dump and then
// applies common labels and
func decodeAndLabelMetrics(encodedMetrics, instanceName, executorVersion string) ([]*dto.MetricFamily, error) {
	data := []*dto.MetricFamily{}

	dec := expfmt.NewDecoder(strings.NewReader(encodedMetrics), expfmt.FmtText)
	for {
		var mf dto.MetricFamily
		if err := dec.Decode(&mf); err != nil {
			if err == io.EOF {
				break
			}

			return nil, errors.Wrap(err, "decoding metric family")
		}

		// Attach the extra labels.
		metricLabelInstance := "instance"
		metricLabelVersion := "version"
		metricLabelJob := "job"
		job := "sourcegraph-executors"
		for _, m := range mf.Metric {
			m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelInstance, Value: &instanceName})
			m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelVersion, Value: &executorVersion})
			m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelJob, Value: &job})
		}

		data = append(data, &mf)
	}

	return data, nil
}
