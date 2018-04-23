package app

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	var errors []string

	if err := db.Ping(r.Context()); err != nil {
		errors = append(errors, "unable to contact PostgreSQL")
	}
	if err := session.Ping(); err != nil {
		errors = append(errors, "unable to contact Redis")
	}

	if len(errors) == 0 {
		fmt.Fprintln(w, "Health check status: OK ✅")
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "Health check status: FAIL ❌")
	fmt.Fprintln(w)
	for _, e := range errors {
		fmt.Fprintln(w, e)
	}
}
