package app

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	var errs []error

	if err := db.Ping(r.Context()); err != nil {
		errs = append(errs, errors.Wrap(err, "PostgreSQL"))
	}
	if err := session.Ping(); err != nil {
		errs = append(errs, errors.Wrap(err, "Redis"))
	}

	if len(errs) == 0 {
		fmt.Fprintln(w, "Health check status: OK ✅")
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "Health check status: FAIL ❌")
	fmt.Fprintln(w)
	for _, err := range errs {
		fmt.Fprintln(w, err)
	}
}
