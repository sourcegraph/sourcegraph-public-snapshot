package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/internal/api"
	proto "github.com/sourcegraph/sourcegraph/internal/api/internalapi/v1"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func serveConfiguration(w http.ResponseWriter, _ *http.Request) error {
	raw := conf.Raw()
	err := json.NewEncoder(w).Encode(raw)
	if err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

// configServer implements proto.ConfigServiceServer to serve config to other clients in the cluster.
type configServer struct {
	proto.UnimplementedConfigServiceServer
}

func (c *configServer) GetConfig(_ context.Context, _ *proto.GetConfigRequest) (*proto.GetConfigResponse, error) {
	raw := conf.Raw()
	return &proto.GetConfigResponse{RawUnified: raw.ToProto()}, nil
}

// gitServiceHandler are handlers which redirect git clone requests to the
// gitserver for the repo.
type gitServiceHandler struct {
	Gitserver interface {
		AddrForRepo(context.Context, api.RepoName) string
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

	addrForRepo := s.Gitserver.AddrForRepo(r.Context(), api.RepoName(repo))
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
