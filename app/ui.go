package app

import (
	"net/http"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveUI(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	if v := os.Getenv("SG_DISABLE_REACTBRIDGE"); v != "" {
		ctx = ui.DisabledReactPrerendering(ctx)
	}

	res, err := ui.RenderRouter(ctx, r, nil)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "ui.html", http.StatusOK, nil, &struct {
		tmpl.Common
		RenderResult *ui.RenderResult
	}{
		RenderResult: res,
	})
}
