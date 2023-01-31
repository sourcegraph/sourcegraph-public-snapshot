package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/grafana/regexp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SetupRoutes registers all route handlers required for all configured executor
// queues with the given router.
func SetupRoutes(handler ExecutorHandler, router *mux.Router) {
	subRouter := router.PathPrefix(fmt.Sprintf("/{queueName:(?:%s)}", regexp.QuoteMeta(handler.Name()))).Subrouter()
	subRouter.Path("/dequeue").Methods(http.MethodPost).HandlerFunc(handler.HandleDequeue)
	subRouter.Path("/heartbeat").Methods(http.MethodPost).HandlerFunc(handler.HandleHeartbeat)
	subRouter.Path("/canceledJobs").Methods(http.MethodPost).HandlerFunc(handler.HandleCanceledJobs)
}

// SetupJobRoutes registers all route handlers required for all configured executor
// queues with the given router.
func SetupJobRoutes(handler ExecutorHandler, router *mux.Router) {
	subRouter := router.PathPrefix(fmt.Sprintf("/{queueName:(?:%s)}", regexp.QuoteMeta(handler.Name()))).Subrouter()
	subRouter.Path("/addExecutionLogEntry").Methods(http.MethodPost).HandlerFunc(handler.HandleAddExecutionLogEntry)
	subRouter.Path("/updateExecutionLogEntry").Methods(http.MethodPost).HandlerFunc(handler.HandleUpdateExecutionLogEntry)
	subRouter.Path("/markComplete").Methods(http.MethodPost).HandlerFunc(handler.HandleMarkComplete)
	subRouter.Path("/markErrored").Methods(http.MethodPost).HandlerFunc(handler.HandleMarkErrored)
	subRouter.Path("/markFailed").Methods(http.MethodPost).HandlerFunc(handler.HandleMarkFailed)
	subRouter.Use(handler.AuthMiddleware)
}

func (h *handler[T]) Name() string {
	return h.queueHandler.Name
}

func (h *handler[T]) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.validateJobRequest(w, r) {
			next.ServeHTTP(w, r)
		}
	})
}

func (h *handler[T]) validateJobRequest(w http.ResponseWriter, r *http.Request) bool {
	// Read the body and re-set the body, so we can parse the request payload.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log15.Error("failed to read request body", "err", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return false
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Every job requests has the basics. Parse out the info we need to see whether the request is valid/authenticated.
	var payload executor.JobOperationRequest
	if err = json.Unmarshal(body, &payload); err != nil {
		log15.Error("failed to parse request body", "err", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return false
	}

	// Since the payload partially deserialize, ensure the worker hostname is valid.
	if err = validateWorkerHostname(payload.ExecutorName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// Get the auth token from the Authorization header.
	var tokenType string
	var authToken string
	if headerValue := r.Header.Get("Authorization"); headerValue != "" {
		parts := strings.Split(headerValue, " ")
		if len(parts) != 2 {
			http.Error(w, fmt.Sprintf(`HTTP Authorization request header value must be of the following form: '%s "TOKEN"' or '%s TOKEN'`, "Bearer", "token-executor"), http.StatusUnauthorized)
			return false
		}
		// Check what the token type is. For backwards compatibility sake, we should also accept the general executor
		// access token.
		tokenType = parts[0]
		if tokenType != "Bearer" && tokenType != "token-executor" {
			http.Error(w, fmt.Sprintf("unrecognized HTTP Authorization request header scheme (supported values: %q, %q)", "Bearer", "token-executor"), http.StatusUnauthorized)
			return false
		}

		authToken = parts[1]
	}
	if authToken == "" {
		http.Error(w, "no token value in the HTTP Authorization request header", http.StatusUnauthorized)
		return false
	}

	// If the general executor access token was provided, simply check the value.
	if tokenType == "token-executor" {
		if authToken == conf.SiteConfig().ExecutorsAccessToken {
			return true
		} else {
			w.WriteHeader(http.StatusForbidden)
			return false
		}
	}

	jobToken, err := h.jobTokenStore.GetByToken(r.Context(), authToken)
	if err != nil {
		log15.Error("failed to retrieve token", "err", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return false
	}

	// Ensure the token was generated for the correct job.
	if jobToken.JobId != int64(payload.JobID) {
		log15.Error("job ID does not match")
		http.Error(w, "invalid token", http.StatusForbidden)
		return false
	}
	// Ensure the token was generated for the correct queue.
	if jobToken.Queue != mux.Vars(r)["queueName"] {
		log15.Error("queue name does not match")
		http.Error(w, "invalid token", http.StatusForbidden)
		return false
	}
	// Ensure the token came from a legit executor instance.
	if _, _, err = h.executorStore.GetByHostname(r.Context(), payload.ExecutorName); err != nil {
		log15.Error("failed to lookup executor by hostname", "err", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return false
	}

	return true
}

// POST /{queueName}/dequeue
func (h *handler[T]) HandleDequeue(w http.ResponseWriter, r *http.Request) {
	var payload executor.DequeueRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		job, dequeued, err := h.dequeue(r.Context(), mux.Vars(r)["queueName"], executorMetadata{
			name:    payload.ExecutorName,
			version: payload.Version,
			resources: ResourceMetadata{
				NumCPUs:   payload.NumCPUs,
				Memory:    payload.Memory,
				DiskSpace: payload.DiskSpace,
			},
		})
		if !dequeued {
			return http.StatusNoContent, nil, err
		}

		return http.StatusOK, job, err
	})
}

// POST /{queueName}/addExecutionLogEntry
func (h *handler[T]) HandleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	var payload executor.AddExecutionLogEntryRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		id, err := h.addExecutionLogEntry(r.Context(), payload.ExecutorName, payload.JobID, payload.ExecutionLogEntry)
		return http.StatusOK, id, err
	})
}

// POST /{queueName}/updateExecutionLogEntry
func (h *handler[T]) HandleUpdateExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	var payload executor.UpdateExecutionLogEntryRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.updateExecutionLogEntry(r.Context(), payload.ExecutorName, payload.JobID, payload.EntryID, payload.ExecutionLogEntry)
		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/markComplete
func (h *handler[T]) HandleMarkComplete(w http.ResponseWriter, r *http.Request) {
	var payload executor.MarkCompleteRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.markComplete(r.Context(), mux.Vars(r)["queueName"], payload.ExecutorName, payload.JobID)
		if err == ErrUnknownJob {
			return http.StatusNotFound, nil, nil
		}

		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/markErrored
func (h *handler[T]) HandleMarkErrored(w http.ResponseWriter, r *http.Request) {
	var payload executor.MarkErroredRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.markErrored(r.Context(), mux.Vars(r)["queueName"], payload.ExecutorName, payload.JobID, payload.ErrorMessage)
		if err == ErrUnknownJob {
			return http.StatusNotFound, nil, nil
		}

		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/markFailed
func (h *handler[T]) HandleMarkFailed(w http.ResponseWriter, r *http.Request) {
	var payload executor.MarkErroredRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.markFailed(r.Context(), mux.Vars(r)["queueName"], payload.ExecutorName, payload.JobID, payload.ErrorMessage)
		if err == ErrUnknownJob {
			return http.StatusNotFound, nil, nil
		}

		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/heartbeat
func (h *handler[T]) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var payload executor.HeartbeatRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		e := types.Executor{
			Hostname:        payload.ExecutorName,
			QueueName:       mux.Vars(r)["queueName"],
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
			metrics, err := decodeAndLabelMetrics(payload.PrometheusMetrics, payload.ExecutorName)
			if err != nil {
				// Just log the error but don't panic. The heartbeat is more important.
				h.logger.Error("failed to decode metrics and apply labels for executor heartbeat", log.Error(err))
				return
			}

			if err := h.metricsStore.Ingest(payload.ExecutorName, metrics); err != nil {
				// Just log the error but don't panic. The heartbeat is more important.
				h.logger.Error("failed to ingest metrics for executor heartbeat", log.Error(err))
			}
		}()

		knownIDs, cancelIDs, err := h.heartbeat(r.Context(), mux.Vars(r)["queueName"], e, payload.JobIDs)

		if payload.Version == executor.ExecutorAPIVersion2 {
			return http.StatusOK, executor.HeartbeatResponse{KnownIDs: knownIDs, CancelIDs: cancelIDs}, err
		}

		// TODO: Remove in Sourcegraph 4.4.
		return http.StatusOK, knownIDs, err
	})
}

// POST /{queueName}/canceledJobs
// TODO: This handler can be removed in Sourcegraph 4.4.
func (h *handler[T]) HandleCanceledJobs(w http.ResponseWriter, r *http.Request) {
	var payload executor.CanceledJobsRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		canceledIDs, err := h.canceled(r.Context(), mux.Vars(r)["queueName"], payload.ExecutorName, payload.KnownJobIDs)
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
func (h *handler[T]) wrapHandler(w http.ResponseWriter, r *http.Request, payload any, handler func() (int, any, error)) {
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
// applies common labels.
func decodeAndLabelMetrics(encodedMetrics, instanceName string) ([]*dto.MetricFamily, error) {
	var data []*dto.MetricFamily

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
		metricLabelInstance := "sg_instance"
		metricLabelJob := "sg_job"
		executorJob := "sourcegraph-executors"
		registryJob := "sourcegraph-executors-registry"
		for _, m := range mf.Metric {
			var metricLabelInstanceValue string
			for _, l := range m.Label {
				if *l.Name == metricLabelInstance {
					metricLabelInstanceValue = l.GetValue()
					break
				}
			}
			// if sg_instance not set, set it as the executor name sent in the heartbeat.
			// this is done for the executor's own and it's node_exporter metrics, executors
			// set sg_instance for metrics scraped from the registry+registry's node_exporter
			if metricLabelInstanceValue == "" {
				m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelInstance, Value: &instanceName})
			}

			if metricLabelInstanceValue == "docker-registry" {
				m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelJob, Value: &registryJob})
			} else {
				m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelJob, Value: &executorJob})
			}
		}

		data = append(data, &mf)
	}

	return data, nil
}
