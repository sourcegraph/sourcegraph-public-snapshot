package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// handleGetGitolitePhabricatorMetadata serves the Gitolite
// Phabricator metadata endpoint, which returns the Phabricator
// metadata for a given repository by running a user-provided command.
func (s *Server) handleGetGitolitePhabricatorMetadata(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := func(name string) bool { _, ok := q[name]; return ok }
	switch {
	case query("gitolite"):
		gitoliteHost := q.Get("gitolite")
		repoName := q.Get("repo")
		for _, gconf := range conf.Get().Gitolite {
			if gconf.Host != gitoliteHost {
				continue
			}
			if gconf.PhabricatorMetadataCommand == "" {
				continue
			}
			callsign, err := getGitolitePhabCallsign(r.Context(), gconf, repoName, gconf.PhabricatorMetadataCommand)
			if err != nil {
				continue
			}

			err = json.NewEncoder(w).Encode(protocol.GitolitePhabricatorMetadataResponse{
				Callsign: callsign,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		// No matching Phabricator host found
		if err := json.NewEncoder(w).Encode(protocol.GitolitePhabricatorMetadataResponse{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "unrecognized URL in /get request", http.StatusNotFound)
		return
	}
}

func getGitolitePhabCallsign(ctx context.Context, gconf schema.GitoliteConnection, repo string, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Env = append(os.Environ(), "REPO="+repo)
	stdout, err := cmd.Output()
	if err != nil {
		log15.Warn("Command to get Gitolite Phabricator callsign failed", "repo", repo, "error", err, "stderr")
		return "", err
	}
	callsign := strings.TrimSpace(string(stdout))
	if callsign == "" {
		return "", fmt.Errorf("callsign command returned empty")
	}
	return callsign, nil
}
