package gitserver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sourcegraph/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
	gitrouter "src.sourcegraph.com/sourcegraph/gitserver/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	httpapiauth "src.sourcegraph.com/sourcegraph/httpapi/auth"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// AddHandlers adds git HTTP handlers to r.
func AddHandlers(r *mux.Router) {
	r.Get(gitrouter.GitInfoRefs).Handler(handler(serveInfoRefs))
	r.Get(gitrouter.GitReceivePack).Handler(handler(serveReceivePack))
	r.Get(gitrouter.GitUploadPack).Handler(handler(serveUploadPack))
}

// handler is a wrapper func for API handlers.
func handler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	mw := []handlerutil.Middleware{
		httpapiauth.PasswordMiddleware,
		httpapiauth.OAuth2AccessTokenMiddleware,
	}
	hh := handlerutil.Handler(handlerutil.HandlerWithErrorReturn{
		Handler: h,
		Error:   handleError,
	})
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
	if handlerutil.DebugMode(r) {
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
	ctx := httpctx.FromRequest(r)

	cred := sourcegraph.CredentialsFromContext(ctx)
	if cred != nil {
		tok, err := cred.Token()
		if err == nil {
			log15.Info("serveInfoRefs creds", "token", tok.AccessToken)
		}
	}

	repo, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	service := trimGitService(r.URL.Query().Get("service"))

	c, err := client(ctx)
	if err != nil {
		return err
	}
	pkt, err := c.InfoRefs(ctx, &gitpb.InfoRefsOp{
		Repo:    repo,
		Service: service,
	})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service))
	noCache(w)
	_, err = w.Write(pkt.Data)
	return err
}

func serveReceivePack(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)

	repo, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	c, err := client(ctx)
	if err != nil {
		return err
	}
	pkt, err := c.ReceivePack(ctx, &gitpb.ReceivePackOp{
		Repo:            repo,
		ContentEncoding: r.Header.Get("content-encoding"),
		Data:            body,
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
	ctx := httpctx.FromRequest(r)

	repo, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	c, err := client(ctx)
	if err != nil {
		return err
	}
	pkt, err := c.UploadPack(ctx, &gitpb.UploadPackOp{
		Repo:            repo,
		ContentEncoding: r.Header.Get("content-encoding"),
		Data:            body,
	})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	noCache(w)
	_, err = w.Write(pkt.Data)
	return err
}

func noCache(w http.ResponseWriter) {
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}
