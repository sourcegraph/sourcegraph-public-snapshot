package appliance

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

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
	_, _ = w.Write(js)

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

func (a *Appliance) getSetupJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func (a *Appliance) postSetupJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Stage string `json:"stage"`
			Data  string `json:"data,omitempty"`
		}

		if err := a.readJSON(w, r, &input); err != nil {
			a.badRequestResponse(w, r, err)
			return
		}
	})
}

func (a *Appliance) getInstallJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func (a *Appliance) getStatusJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Status string
		}{
			Status: a.status.String(),
		}

		if err := a.writeJSON(w, http.StatusOK, responseData{"status": data}, nil); err != nil {
			a.serverErrorResponse(w, r, err)
		}
	})
}
