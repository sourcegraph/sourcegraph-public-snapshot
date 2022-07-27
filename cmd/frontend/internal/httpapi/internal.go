package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// serveExternalServiceConfigs serves a JSON response that is an array of all
// external service configs that match the requested kind.
func serveExternalServiceConfigs(db database.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req api.ExternalServiceConfigsRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return err
		}

		options := database.ExternalServicesListOptions{
			Kinds:   []string{req.Kind},
			AfterID: int64(req.AfterID),
		}
		if req.Limit > 0 {
			options.LimitOffset = &database.LimitOffset{
				Limit: req.Limit,
			}
		}

		services, err := db.ExternalServices().List(r.Context(), options)
		if err != nil {
			return err
		}

		// Instead of returning an intermediate response type, we directly return
		// the array of configs (which are themselves JSON objects).
		// This makes it possible for the caller to directly unmarshal the response into
		// a slice of connection configurations for this external service kind.
		configs := make([]map[string]any, 0, len(services))
		for _, service := range services {
			var config map[string]any
			// Raw configs may have comments in them so we have to use a json parser
			// that supports comments in json.
			if err := jsonc.Unmarshal(service.Config, &config); err != nil {
				log15.Error(
					"ignoring external service config that has invalid json",
					"id", service.ID,
					"displayName", service.DisplayName,
					"config", service.Config,
					"err", err,
				)
				continue
			}
			configs = append(configs, config)
		}
		return json.NewEncoder(w).Encode(configs)
	}
}

func serveConfiguration(w http.ResponseWriter, _ *http.Request) error {
	raw := conf.Raw()
	err := json.NewEncoder(w).Encode(raw)
	if err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveExternalURL(w http.ResponseWriter, _ *http.Request) error {
	if err := json.NewEncoder(w).Encode(globals.ExternalURL().String()); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveSendEmail(_ http.ResponseWriter, r *http.Request) error {
	var msg txemail.Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		return err
	}
	return txemail.Send(r.Context(), msg)
}

// gitServiceHandler are handlers which redirect git clone requests to the
// gitserver for the repo.
type gitServiceHandler struct {
	Gitserver interface {
		AddrForRepo(context.Context, api.RepoName) (string, error)
	}
}

func (s *gitServiceHandler) serveInfoRefs() func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		return s.redirectToGitServer(w, r, "/info/refs")
	}
}

func (s *gitServiceHandler) serveGitUploadPack() func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		return s.redirectToGitServer(w, r, "/git-upload-pack")
	}
}

func (s *gitServiceHandler) redirectToGitServer(w http.ResponseWriter, r *http.Request, gitPath string) error {
	repo := mux.Vars(r)["RepoName"]

	addrForRepo, err := s.Gitserver.AddrForRepo(r.Context(), api.RepoName(repo))
	if err != nil {
		return err
	}
	u := &url.URL{
		Scheme:   "http",
		Host:     addrForRepo,
		Path:     path.Join("/git", repo, gitPath),
		RawQuery: r.URL.RawQuery,
	}

	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
	return nil
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "could not parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, _ = w.Write([]byte("pong"))
}
