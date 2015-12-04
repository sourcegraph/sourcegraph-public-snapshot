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

	if rr.Code == http.StatusUnauthorized && nil == handlerutil.UserFromContext(ctx) {
		return grpc.Errorf(codes.Unauthenticated, "platform search return unauthorized and no authenticated user in current context")
	} else if rr.Code != http.StatusOK {
		// NOTE The internal.Handler handles an error by returning
		// specific error html. We don't want to do that in this case
		// and instead just pass the raw bytes returned from the search
		// frame. This will forward the response body and error code
		// to the client, which allows the search frame to define the
		// desired HTTP semantics.
		w.WriteHeader(rr.Code)
	}

	_, err = io.Copy(w, rr.Body)
	return err
}
