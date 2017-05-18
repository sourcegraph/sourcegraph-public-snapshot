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

	for _, entry := range gitoliteHostMap {
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

	if err := json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
