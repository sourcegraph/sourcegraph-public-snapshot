package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
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
			rawConfig, err := service.Config.Decrypt(r.Context())
			if err != nil {
				return err
			}
			if jsonc.Unmarshal(rawConfig, &config); err != nil {
				log15.Error(
					"ignoring external service config that has invalid json",
					"id", service.ID,
					"displayName", service.DisplayName,
					"config", rawConfig,
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

func decodeSendEmail(r *http.Request) (txtypes.InternalAPIMessage, error) {
	var msg txtypes.InternalAPIMessage
	return msg, json.NewDecoder(r.Body).Decode(&msg)
}

func serveSendEmail(_ http.ResponseWriter, r *http.Request) error {
	msg, err := decodeSendEmail(r)
	if err != nil {
		return errors.Wrap(err, "decode request")
	}
	return txemail.Send(r.Context(), msg.Source, msg.Message)
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

func newServiceRegisterHandler(db database.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client := requestclient.FromContext(r.Context())
		if client == nil {
			http.Error(w, "could not extract IP address", http.StatusBadRequest)
			return
		}

		ip, err := netip.ParseAddr(client.IP)
		if err != nil {
			http.Error(w, fmt.Sprintf("cloud not parse client IP %s: %s", client.IP, err.Error()), http.StatusBadRequest)
			return
		}

		args := database.ServiceArgs{}

		err = json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check required args.
		if args.Port == 0 {
			http.Error(w, "missing port", http.StatusBadRequest)
			return
		}

		if args.Hostname == "" {
			http.Error(w, "missing hostname", http.StatusBadRequest)
		}

		// Hydrate ip. We extracted the ip from the request headers and not the body.
		args.IP = ip

		vars := mux.Vars(r)
		id, err := db.Services().Register(r.Context(), vars["name"], args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return ID to caller.
		w.Write([]byte(id))
	}
}

func newServiceRenewHandler(db database.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		err := db.Services().Renew(r.Context(), vars["name"], vars["instanceID"])
		if err != nil {
			if errcode.IsNotFound(err) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
}

func newServiceDeregisterHandler(db database.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		err := db.Services().Deregister(r.Context(), vars["name"], vars["instanceID"])
		if err != nil {
			if err != nil {
				if errcode.IsNotFound(err) {
					http.Error(w, err.Error(), http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}
	}
}
