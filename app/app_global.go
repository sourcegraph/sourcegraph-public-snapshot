package app

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	gcontext "github.com/gorilla/context"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// TOOD: This code is very similar in 3 places: here, serveRepoPlatformSearchResults and in serveRepoFrame.
//       Consider refactoring the common parts out.
func serveAppGlobalNotificationCenter(w http.ResponseWriter, r *http.Request) error {
	// TODO(beyang): think of more robust way of isolating apps to
	// prevent shared mutable state (e.g., modifying http.Requests) to
	// prevent inter-app interference
	rCopy := *r
	urlCopy := *r.URL
	rCopy.URL = &urlCopy

	ctx := httpctx.FromRequest(r)

	appctx, err := pctx.WithGlobalAppInfo(ctx, r)
	if err != nil {
		return err
	}
	httpctx.SetForRequest(&rCopy, appctx)
	defer gcontext.Clear(&rCopy) // clear the app context after finished to avoid a memory leak

	rr := httptest.NewRecorder()

	stripPrefix := pctx.BaseURI(appctx)
	if u, err := url.Parse(stripPrefix); err == nil {
		stripPrefix = u.Path
	} else {
		return err
	}

	// The canonical URL for app root page does not have a trailing slash, so redirect.
	if rCopy.URL.Path == stripPrefix+"/" {
		baseURL := stripPrefix
		if rCopy.URL.RawQuery != "" {
			baseURL += "?" + rCopy.URL.RawQuery
		}
		http.Redirect(w, r, baseURL, http.StatusMovedPermanently)
		return nil
	}

	// strip prefix
	if p := strings.TrimPrefix(rCopy.URL.Path, stripPrefix); len(p) < len(r.URL.Path) {
		rCopy.URL.Path = p
		if rCopy.URL.Path == "" { // For the app http.Handler, the root path should always be "/".
			rCopy.URL.Path = "/"
		}
	} else {
		return fmt.Errorf("could not load app: %q was not a prefix of %q", stripPrefix, rCopy.URL.Path)
	}

	handler := platform.GlobalApps["notifications"].Handler
	handler.ServeHTTP(rr, &rCopy)

	// extract response body (purposefully ignoring headers)
	body := string(rr.Body.Bytes())

	// If Sourcegraph-Verbatim header was set to true, or this is a redirect,
	// relay this request to browser directly, and copy appropriate headers.
	redirect := rr.Code == http.StatusSeeOther || rr.Code == http.StatusMovedPermanently || rr.Code == http.StatusTemporaryRedirect || rr.Code == http.StatusFound
	if rr.Header().Get(platform.HTTPHeaderVerbatim) == "true" || redirect {
		copyHeader(w.Header(), rr.Header())
		w.WriteHeader(rr.Code)
		_, err := io.Copy(w, rr.Body)
		return err
	}

	var appHTML template.HTML
	var appError error
	if rr.Code == http.StatusOK {
		appHTML = template.HTML(body)
	} else if rr.Code == http.StatusUnauthorized && nil == handlerutil.UserFromContext(ctx) {
		// App returned Unauthorized, and user's not logged in. So redirect to login page and try again.
		return grpc.Errorf(codes.Unauthenticated, "platform app returned unauthorized and no authenticated user in current context")
	} else {
		appError = errors.New(body)
	}

	tmplData := struct {
		AppHTML  template.HTML
		AppError error

		tmpl.Common
	}{
		AppHTML:  appHTML,
		AppError: appError,
	}

	return tmpl.Exec(r, w, "app_global.html", http.StatusOK, nil, &tmplData)
}
