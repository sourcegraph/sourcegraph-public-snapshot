package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	gcontext "github.com/gorilla/context"

	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"

	"github.com/sourcegraph/mux"
)

// serveAppSearchResults fetches the map of all SearchFrames registered
// through the platform package. It also adds the base URL  path of the request
// and a CSRF token to the request context so that platform apps can access them.
func serveRepoPlatformSearchResults(w http.ResponseWriter, r *http.Request) error {
	appID := mux.Vars(r)["AppID"]

	searchFrames := platform.SearchFrames()

	searchFrame, ok := searchFrames[appID]
	if !ok {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("Search frame %q was not found", appID),
		}
	}

	rCopy := copyRequest(r)

	ctx := httpctx.FromRequest(r)

	framectx, err := pctx.WithRepoSearchInfo(ctx, r)
	if err != nil {
		return err
	}

	httpctx.SetForRequest(rCopy, framectx)
	defer gcontext.Clear(rCopy)

	platform.SetPlatformRequestURL(framectx, w, r, rCopy)

	rr := httptest.NewRecorder()

	searchFrame.Handler.ServeHTTP(rr, rCopy)

	if rr.Code == http.StatusOK {
		_, err := io.Copy(w, rr.Body)
		return err
	} else if rr.Code == http.StatusUnauthorized && nil == handlerutil.UserFromContext(ctx) {
		return grpc.Errorf(codes.Unauthenticated, "platform search return unauthorized and no authenticated user in current context")
	} else {
		// TODO(poler) Should other response codes have specific semantics?
		return fmt.Errorf("Unexpected response code from %q search frame: %d", searchFrame.ID, rr.Code)
	}
}

func copyRequest(r *http.Request) *http.Request {
	rCopy := *r
	urlCopy := *r.URL
	rCopy.URL = &urlCopy
	return &rCopy
}
