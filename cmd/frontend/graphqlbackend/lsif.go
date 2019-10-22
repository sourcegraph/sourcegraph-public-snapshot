package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"golang.org/x/net/context/ctxhttp"
)

var lsifServerURLFromEnv = "http://localhost:3186" // TODO

func lsifRequest(ctx context.Context, path string, query url.Values, payload interface{}) (err error) {
	tr, ctx := trace.New(ctx, "lsifRequest", fmt.Sprintf("path: %s", path))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	u, err := url.Parse(fmt.Sprintf("%s/%s", lsifServerURLFromEnv, path))
	if err != nil {
		return err
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("LSIF client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	// TODO(efritz): use a separate HTTP client for LSIF so we do
	// not have the same limits as search/codemod?
	resp, err := ctxhttp.Do(ctx, searchHTTPClient, req)
	if err != nil {
		// If we failed due to cancellation or timeout (with no partial results in the response body), return just that.
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return errors.Wrap(err, "lsif request failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.WithStack(&lsifError{StatusCode: resp.StatusCode, Message: string(body)})
	}

	err = json.Unmarshal([]byte(body), &payload)
	return nil
}

type lsifError struct {
	StatusCode int
	Message    string
}

func (e *lsifError) Error() string {
	return e.Message
}
