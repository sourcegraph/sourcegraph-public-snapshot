package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/gorilla/mux"
	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/log"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SetupRoutes registers all route handlers required for all configured executor
// queues with the given router.
func SetupRoutes(handlers []ExecutorHandler, router *mux.Router) {
	for _, h := range handlers {
		subRouter := router.PathPrefix(fmt.Sprintf("/{queueName:(?:%s)}/", regexp.QuoteMeta(h.Name()))).Subrouter()
		routes := map[string]func(w http.ResponseWriter, r *http.Request){
			"dequeue":   h.handleDequeue,
			"heartbeat": h.handleHeartbeat,
			// TODO: This endpoint can be removed in Sourcegraph 4.4.
			"canceledJobs": h.handleCanceledJobs,
		}
		for path, route := range routes {
			subRouter.Path(fmt.Sprintf("/%s", path)).Methods("POST").HandlerFunc(route)
		}
	}
}

// SetupJobRoutes registers all route handlers required for all configured executor
// queues with the given router.
func SetupJobRoutes(handlers []ExecutorHandler, router *mux.Router) {
	for _, h := range handlers {
		subRouter := router.PathPrefix(fmt.Sprintf("/{queueName:(?:%s)}/", regexp.QuoteMeta(h.Name()))).Subrouter()
		routes := map[string]func(w http.ResponseWriter, r *http.Request){
			"addExecutionLogEntry":    h.handleAddExecutionLogEntry,
			"updateExecutionLogEntry": h.handleUpdateExecutionLogEntry,
			"markComplete":            h.handleMarkComplete,
			"markErrored":             h.handleMarkErrored,
			"markFailed":              h.handleMarkFailed,
		}
		for path, route := range routes {
			subRouter.Path(fmt.Sprintf("/%s", path)).Methods("POST").HandlerFunc(route)
		}
	}
}

// JobAuthMiddleware authenticates job requests,
func JobAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validateJobRequest(w, r) {
			next.ServeHTTP(w, r)
		}
	})
}

func validateJobRequest(w http.ResponseWriter, r *http.Request) bool {
	// Read the body and re-set the body, so we can parse the request payload.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log15.Error("failed to read request body", "err", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return false
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Every job requests has the basics. Parse out the info we need to see whether the request is valid/authenticated.
	var payload apiclient.JobOperationRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		log15.Error("failed to parse request body", "err", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
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

	accessToken := conf.SiteConfig().ExecutorsAccessToken

	// If the general executor access token was provided, simply check the value.
	if tokenType == "token-executor" {
		if authToken == accessToken {
			return true
		} else {
			w.WriteHeader(http.StatusForbidden)
			return false
		}
	}

	// Parse the provided JWT token with the key.
	c := jobOperationClaims{}
	token, err := jwt.ParseWithClaims(authToken, &c, func(token *jwt.Token) (any, error) {
		return base64.StdEncoding.DecodeString(conf.SiteConfig().Executors.JobAccessToken.SigningKey)
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Name}))

	if err != nil {
		log15.Error("failed to parse token", "err", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return false
	}

	// Time to go to work on the provided token.
	if claims, ok := token.Claims.(*jobOperationClaims); ok && token.Valid {
		// Make sure the request has the general executor access token.
		if claims.AccessToken != accessToken {
			log15.Error("claims access token does not match the executor access token")
			w.WriteHeader(http.StatusForbidden)
			return false
		}
		// Ensure the token the request was generated for is coming for the correct executor instance.
		if claims.Issuer != payload.ExecutorName {
			log15.Error("executor name does not match claims Issuer")
			w.WriteHeader(http.StatusForbidden)
			return false
		}
		// Parse the job ID from the claims.
		id, err := strconv.Atoi(claims.Subject)
		if err != nil {
			log15.Error("failed to claim Subject to integer", "err", err)
			w.WriteHeader(http.StatusForbidden)
			return false
		}
		// Ensure the job matches the payload
		if id != payload.JobID {
			log15.Error("job ID does not match claims Subject")
			w.WriteHeader(http.StatusForbidden)
			return false
		}
	}

	// Since the payload partially deserialize, ensure the worker hostname is valid.
	if err = validateWorkerHostname(payload.ExecutorName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

// POST /{queueName}/dequeue
func (h *handler[T]) handleDequeue(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.DequeueRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		job, dequeued, err := h.dequeue(r.Context(), executorMetadata{
			Name:    payload.ExecutorName,
			Version: payload.Version,
			Resources: ResourceMetadata{
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
func (h *handler[T]) handleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.AddExecutionLogEntryRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		id, err := h.addExecutionLogEntry(r.Context(), payload.ExecutorName, payload.JobID, payload.ExecutionLogEntry)
		return http.StatusOK, id, err
	})
}

// POST /{queueName}/updateExecutionLogEntry
func (h *handler[T]) handleUpdateExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.UpdateExecutionLogEntryRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		err := h.updateExecutionLogEntry(r.Context(), payload.ExecutorName, payload.JobID, payload.EntryID, payload.ExecutionLogEntry)
		return http.StatusNoContent, nil, err
	})
}

// POST /{queueName}/markComplete
func (h *handler[T]) handleMarkComplete(w http.ResponseWriter, r *http.Request) {
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
func (h *handler[T]) handleMarkErrored(w http.ResponseWriter, r *http.Request) {
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
func (h *handler[T]) handleMarkFailed(w http.ResponseWriter, r *http.Request) {
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
func (h *handler[T]) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
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

		knownIDs, cancelIDs, err := h.heartbeat(r.Context(), executor, payload.JobIDs)

		if payload.Version == apiclient.ExecutorAPIVersion2 {
			return http.StatusOK, apiclient.HeartbeatResponse{KnownIDs: knownIDs, CancelIDs: cancelIDs}, err
		}

		// TODO: Remove in Sourcegraph 4.4.
		return http.StatusOK, knownIDs, err
	})
}

// POST /{queueName}/canceledJobs
// TODO: This handler can be removed in Sourcegraph 4.4.
func (h *handler[T]) handleCanceledJobs(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.CanceledJobsRequest

	h.wrapHandler(w, r, &payload, func() (int, any, error) {
		canceledIDs, err := h.canceled(r.Context(), payload.ExecutorName, payload.KnownJobIDs)
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
