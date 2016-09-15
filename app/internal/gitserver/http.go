package gitserver

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"context"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	gitrouter "sourcegraph.com/sourcegraph/sourcegraph/app/internal/gitserver/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/auth"
)

// AddHandlers adds git HTTP handlers to r.
func AddHandlers(r *mux.Router) {
	r.Get(gitrouter.GitInfoRefs).Handler(httptrace.TraceRoute(handler(serveInfoRefs)))
	r.Get(gitrouter.GitReceivePack).Handler(httptrace.TraceRoute(handler(serveReceivePack)))
	r.Get(gitrouter.GitUploadPack).Handler(httptrace.TraceRoute(handler(serveUploadPack)))
}

// handler is a wrapper func for API handlers.
func handler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	mw := []handlerutil.Middleware{
		httpapiauth.VaryHeader,
		httpapiauth.PasswordMiddleware,
		httpapiauth.OAuth2AccessTokenMiddleware,
	}
	hh := handlerutil.HandlerWithErrorReturn{
		Handler: h,
		Error:   handleError,
	}
	return handlerutil.WithMiddleware(hh, mw...)
}

func handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	// Never cache error responses.
	w.Header().Set("cache-control", "no-cache, max-age=0")

	if status == http.StatusUnauthorized {
		// git needs this header to attempt authentication.
		w.Header().Set("www-authenticate", `Basic realm="git repository"`)
	}

	errBody := err.Error()

	var displayErrBody string
	if handlerutil.DebugMode {
		// Only display error message to admins or locally, since it
		// can contain sensitive info (like API keys in net/http error
		// messages).
		displayErrBody = string(errBody)
	}
	http.Error(w, displayErrBody, status)
	if status < 200 || status >= 500 {
		log15.Debug("gitserver.handleError called with unsuccessful status code", "method", r.Method, "request_uri", r.URL.RequestURI(), "status_code", status, "error", err.Error())
	}
}

func trimGitService(name string) string {
	return strings.TrimSpace(strings.TrimPrefix(name, "git-"))
}

func serveInfoRefs(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	repo, err := getRepoID(ctx, routevar.ToRepo(mux.Vars(r)))
	if err != nil {
		return err
	}

	service := trimGitService(r.URL.Query().Get("service"))

	c, err := client(ctx)
	if err != nil {
		return err
	}

	var pkt *sourcegraph.Packet
	switch service {
	case "receive-pack":
		var err error
		pkt, err = c.ReceivePack(ctx, &sourcegraph.ReceivePackOp{
			Repo:          repo,
			AdvertiseRefs: true,
		})
		if err != nil {
			return err
		}
	case "upload-pack":
		var err error
		pkt, err = c.UploadPack(ctx, &sourcegraph.UploadPackOp{
			Repo:          repo,
			AdvertiseRefs: true,
		})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unrecognized git service: %q", service)
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service))
	noCache(w)
	if _, err := fmt.Fprintf(w, "%04x# service=git-%s\n0000", 19+len(service), service); err != nil {
		return err
	}
	_, err = w.Write(pkt.Data)
	return err
}

func serveReceivePack(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	repo, err := getRepoID(ctx, routevar.ToRepo(mux.Vars(r)))
	if err != nil {
		return err
	}

	body, err := readBody(r.Body, r.Header.Get("content-encoding"))
	if err != nil {
		return err
	}

	c, err := client(ctx)
	if err != nil {
		return err
	}
	pkt, err := c.ReceivePack(ctx, &sourcegraph.ReceivePackOp{
		Repo: repo,
		Data: body,
	})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	noCache(w)
	_, err = w.Write(pkt.Data)
	return err
}

func serveUploadPack(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	repo, err := getRepoID(ctx, routevar.ToRepo(mux.Vars(r)))
	if err != nil {
		return err
	}

	body, err := readBody(r.Body, r.Header.Get("content-encoding"))
	if err != nil {
		return err
	}

	c, err := client(ctx)
	if err != nil {
		return err
	}
	pkt, err := c.UploadPack(ctx, &sourcegraph.UploadPackOp{
		Repo: repo,
		Data: body,
	})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	noCache(w)
	_, err = w.Write(pkt.Data)
	return err
}

func getRepoID(ctx context.Context, repoPath string) (int32, error) {
	c, err := client(ctx)
	if err != nil {
		return 0, err
	}

	res, err := c.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repoPath})
	if err != nil {
		return 0, err
	}
	return res.Repo, nil
}

func noCache(w http.ResponseWriter) {
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}

func readBody(body io.Reader, encoding string) ([]byte, error) {
	switch encoding {
	case "gzip":
		gr, err := gzip.NewReader(body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		return ioutil.ReadAll(gr)

	case "deflate":
		fr := flate.NewReader(body)
		defer fr.Close()
		return ioutil.ReadAll(fr)

	case "":
		return ioutil.ReadAll(body)

	default:
		return nil, fmt.Errorf("unrecognized git content encoding: %q", encoding)
	}
}
