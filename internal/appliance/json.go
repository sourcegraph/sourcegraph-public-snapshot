package appliance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const maxBytes = 1_048_576

type responseData map[string]any

func (a *Appliance) writeJSON(w http.ResponseWriter, status int, data responseData, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}

func (a *Appliance) readJSON(w http.ResponseWriter, r *http.Request, output any) error {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(output)
	if err != nil {
		var jsonMaxBytesErrorType *http.MaxBytesError
		var jsonSyntaxErrorType *json.SyntaxError
		var jsonUnmarshalErrorType *json.UnmarshalTypeError
		var jsonInvalidUnmarshalErrorType *json.InvalidUnmarshalError

		// list of de-facto errors common to JSON APIs that we want to wrap and handle
		switch {
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			return errors.Newf("request body contains unknown key")

		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("malformed JSON contained in request body")

		case errors.As(err, &jsonSyntaxErrorType):
			return errors.Newf("malformed JSON found at character %d", jsonSyntaxErrorType.Offset)

		case errors.As(err, &jsonMaxBytesErrorType):
			return errors.Newf("request body larger than %d bytes", jsonMaxBytesErrorType.Limit)

		case errors.As(err, &jsonUnmarshalErrorType):
			if jsonUnmarshalErrorType.Field != "" {
				return errors.Newf("incorrect JSON type for field %q", jsonUnmarshalErrorType.Field)
			}
			return errors.Newf("incorrect JSON type found at character %d", jsonUnmarshalErrorType.Offset)

		case errors.As(err, &jsonInvalidUnmarshalErrorType):
			panic(err)

		default:
			return err
		}
	}

	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("request body must only contain single JSON value")
	}

	return nil
}

func (a *Appliance) getStatusJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Status string `json:"status"`
			Data   string `json:"data,omitempty"`
		}{
			Status: a.status.String(),
			Data:   "",
		}

		if err := a.writeJSON(w, http.StatusOK, responseData{"status": data}, http.Header{}); err != nil {
			a.serverErrorResponse(w, r, err)
		}
	})
}

func (a *Appliance) getInstallProgressJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentTasks, progress := calculateProgress(installTasks())

		installProgress := struct {
			Version  string `json:"version"`
			Progress int    `json:"progress"`
			Error    string `json:"error"`
			Tasks    []Task `json:"tasks"`
		}{
			Version:  "",
			Progress: progress,
			Error:    "",
			Tasks:    currentTasks,
		}

		ok, err := a.isSourcegraphFrontendReady(r.Context())
		if err != nil {
			a.logger.Error("failed to get sourcegraph frontend status")
			return
		}

		if ok {
			a.status = config.StatusWaitingForAdmin
		}

		if err := a.writeJSON(w, http.StatusOK, responseData{"progress": installProgress}, http.Header{}); err != nil {
			a.serverErrorResponse(w, r, err)
		}
	})
}

func (a *Appliance) getMaintenanceStatusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		type service struct {
			Name    string `json:"name"`
			Healthy bool   `json:"healthy"`
			Message string `json:"message"`
		}

		services := []service{}
		for _, name := range config.SourcegraphServicesToReconcile {
			services = append(services, service{
				Name:    name,
				Healthy: true,
				Message: "fake event",
			})
		}
		fmt.Println(services)
		if err := a.writeJSON(w, http.StatusOK, responseData{"services": services}, http.Header{}); err != nil {
			a.serverErrorResponse(w, r, err)
		}
	})
}

func (a *Appliance) getReleasesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var relregResp io.ReadCloser
		if a.pinnedReleasesFile == "" {
			// simple proxy to release releaseRegistry
			resp, err := http.Get("https://releaseregistry.sourcegraph.com/v1/releases/sourcegraph")
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}
			defer resp.Body.Close()
			relregResp = resp.Body
		} else {
			// airgap fallback
			var err error
			relregResp, err = os.Open(a.pinnedReleasesFile)
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}
			defer relregResp.Close()
		}

		if _, err := io.Copy(w, relregResp); err != nil {
			// There's nothing else we can do, we've already sent the status
			// code
			a.logger.Error("error proxying release registry to appliance frontend", log.Error(err))
		}
	})
}

func (a *Appliance) postStatusJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			State string `json:"state"`
			Data  string `json:"data,omitempty"`
		}

		if err := a.readJSON(w, r, &input); err != nil {
			a.badRequestResponse(w, r, err)
			return
		}

		newStatus := config.Status(input.State)
		a.logger.Info("state transition", log.String("state", string(newStatus)))
		// trim v if v exists
		input.Data = strings.TrimPrefix(input.Data, "v")
		a.sourcegraph.Spec.RequestedVersion = input.Data
		if err := a.setStatus(r.Context(), newStatus); err != nil {
			if kerrors.IsNotFound(err) {
				a.logger.Info("no configmap found, will not set status")
			} else {
				a.serverErrorResponse(w, r, err)
				return
			}
		}

		if a.noResourceRestrictions {
			a.sourcegraph.SetLocalDevMode()
		}

		cfgMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sourcegraph-appliance",
				Namespace: a.namespace,
			},
		}
		err := a.reconcileConfigMap(r.Context(), cfgMap)
		if err != nil {
			a.serverErrorResponse(w, r, err)
		}

		a.status = newStatus
	})
}
