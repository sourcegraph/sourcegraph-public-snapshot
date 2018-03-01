package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	repos := make([]string, 0)

	q := r.URL.Query()
	query := func(name string) bool { _, ok := q[name]; return ok }
	switch {
	case r.URL.RawQuery == "":
		fallthrough // treat same as if the URL query was "gitolite" for backcompat
	case query("gitolite"):
		for _, entry := range originMaps.getGitoliteHostMap() {
			out, err := exec.CommandContext(r.Context(), "ssh", entry.Origin, "info").CombinedOutput()
			if err != nil {
				log.Printf("listing gitolite failed: %s (Output: %q)", err, string(out))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[0] == "R" {
					name := fields[len(fields)-1]
					repos = append(repos, entry.Prefix+name)
				}
			}
		}

	default:
		// empty list response for unrecognized URL query
	}

	if err := json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
