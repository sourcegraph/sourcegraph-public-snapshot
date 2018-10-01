package server

import (
	"compress/gzip"
	"net/http"
	"os/exec"
	"path"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func (s *Server) handleUploadPack(w http.ResponseWriter, r *http.Request) {
	repo := protocol.NormalizeRepo(api.RepoURI(r.URL.Query().Get("repo")))
	if repo == "" {
		http.Error(w, "repo missing", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/x-git-upload-pack-request" {
		http.Error(w, "Unexpected Content-Type", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")

	body := r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	defer body.Close()

	cmd := exec.CommandContext(r.Context(), "git", "upload-pack", "--stateless-rpc", ".")
	cmd.Dir = path.Join(s.ReposDir, string(repo))
	cmd.Stdout = w
	cmd.Stdin = body
	if err := cmd.Run(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
