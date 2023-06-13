package cli

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func makeUpgradeProgressHandler(db database.DB) http.HandlerFunc {
	// TODO(efritz) - persist plan + progress
	// TODO(efritz) - query plan and progress for display

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var value int
		if err := func() (err error) {
			value, _, err = basestore.ScanFirstInt(db.QueryContext(ctx, `SELECT 4`))
			return err
		}(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, upgradeProgressHandlerTemplate, value)
	}
}

const upgradeProgressHandlerTemplate = `
<body>
	<h1>FANCY MIGRATION IN PROGRESS: %d</h1>
</body>
`
