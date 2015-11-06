package issues

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"strconv"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/sourcegraph/mux"
	netctx "golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/platform/apps/issues/assets"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
)

func unmarshalIssueSpec(r *http.Request) IssueSpec {
	ctx := putil.Context(r)

	var spec IssueSpec
	if repoRevSpec, exists := pctx.RepoRevSpec(ctx); exists {
		spec.Repo = repoRevSpec.URI
	}
	if issueNum, err := strconv.ParseInt(mux.Vars(r)["issueNum"], 10, 64); err == nil {
		spec.ID = issueNum
	}
	return spec
}

func unmarshalReplySpec(r *http.Request) ReplySpec {
	var spec ReplySpec
	spec.IssueSpec = unmarshalIssueSpec(r)
	spec.ReplyNum, _ = strconv.Atoi(mux.Vars(r)["replyNum"])

	return spec
}

func serveMainPage(w http.ResponseWriter, r *http.Request) error {
	spec := unmarshalIssueSpec(r)
	ctx := putil.Context(r)
	sg := sourcegraph.NewClientFromContext(ctx)

	issues, err := Issues.List(spec.Repo)
	if err != nil {
		return err
	}
	for _, issue := range issues.Issues {
		issue.Author, err = sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
		if err != nil {
			return err
		}
	}
	open := issues.Open()
	closed := issues.Closed()

	csrfToken := pctx.CSRFToken(ctx)

	return parse("main.html").Execute(w, struct {
		BaseURI   string
		Closed    IssueList
		Open      IssueList
		CSRFToken string
	}{
		BaseURI:   pctx.BaseURI(ctx),
		Open:      open,
		Closed:    closed,
		CSRFToken: csrfToken,
	})
}

func serveIssue(w http.ResponseWriter, r *http.Request) error {
	spec := unmarshalIssueSpec(r)
	issue, err := Issues.Get(spec)
	if err != nil {
		return err
	}

	ctx := putil.Context(r)
	usr := putil.UserFromRequest(r)

	csrfToken := pctx.CSRFToken(ctx)

	w.Header().Set("X-Sourcegraph-Title", fmt.Sprintf("Issue #%d", issue.UID))

	return parse("issue.html").Execute(w, struct {
		BaseURI   string
		Issue     *Issue
		Creator   bool
		Ctx       netctx.Context
		CSRFToken string
	}{
		BaseURI:   pctx.BaseURI(ctx),
		Issue:     issue,
		Creator:   issue.AuthorUID == usr.UID,
		Ctx:       ctx,
		CSRFToken: csrfToken,
	})
}

func serveNewIssue(w http.ResponseWriter, r *http.Request) error {
	ctx := putil.Context(r)
	usr := putil.UserFromRequest(r)

	spec := unmarshalIssueSpec(r)
	newIssue, err := Issues.Upsert(spec.Repo, &Issue{
		issueInternal: issueInternal{
			Title:     r.PostFormValue("title"),
			Status:    "open",
			AuthorUID: usr.UID,
			Events: []Event{
				Event{
					Created:   time.Now(),
					UID:       1,
					AuthorUID: usr.UID,
					Body:      r.PostFormValue("content"),
				},
			},
		},
	})

	if err != nil {
		return err
	}
	baseURI := pctx.BaseURI(ctx)
	newURL := fmt.Sprintf("%s/%d", baseURI, newIssue.UID)
	passThrough(w)
	http.Redirect(w, r, newURL, http.StatusSeeOther)
	return nil
}

func serveIssueReplyCreate(w http.ResponseWriter, r *http.Request) error {
	spec := unmarshalReplySpec(r)
	ctx := putil.Context(r)
	usr := putil.UserFromRequest(r)

	_, err := Issues.CreateReply(spec.IssueSpec, r.PostFormValue("reply"), usr.UID)
	if err != nil {
		return err
	}

	baseURI := pctx.BaseURI(ctx)
	newURL := fmt.Sprintf("%s/%d", baseURI, spec.ID)
	passThrough(w)
	http.Redirect(w, r, newURL, http.StatusSeeOther)
	return nil
}

func serveCloseIssue(w http.ResponseWriter, r *http.Request) error {
	spec := unmarshalIssueSpec(r)
	ctx := putil.Context(r)
	usr := putil.UserFromRequest(r)

	issue, err := Issues.Get(spec)
	if err != nil {
		return err
	}
	if issue.AuthorUID != usr.UID {
		return ErrNotAuthorized
	}
	issue.Status = "closed"
	Issues.Upsert(spec.Repo, issue)

	newURL := pctx.BaseURI(ctx) + "/"
	passThrough(w)
	http.Redirect(w, r, newURL, http.StatusSeeOther)
	return nil
}

func serveRawReply(w http.ResponseWriter, r *http.Request) error {
	spec := unmarshalIssueSpec(r)
	issue, err := Issues.Get(spec)
	if err != nil {
		return err
	}
	replyNum, err := strconv.Atoi(mux.Vars(r)["replyNum"])
	if err != nil {
		return err
	}

	reply := issue.GetEvent(replyNum)

	passThrough(w)
	w.Write([]byte(reply.Body))
	return nil
}

func serveEditReply(w http.ResponseWriter, r *http.Request) error {
	spec := unmarshalIssueSpec(r)
	ctx := putil.Context(r)

	replyNum, err := strconv.Atoi(mux.Vars(r)["replyNum"])
	if err != nil {
		return err
	}
	issue, err := Issues.UpdateReply(spec, replyNum, r.PostFormValue("reply"))

	baseURI := pctx.BaseURI(ctx)
	newUrl := fmt.Sprintf("%s/%d#%d", baseURI, issue.UID, replyNum)
	passThrough(w)
	http.Redirect(w, r, newUrl, http.StatusSeeOther)
	return nil
}

// TODO(beyang): Is this necessary? What do we need to passthrough besides redirects?
func passThrough(w http.ResponseWriter) {
	w.Header().Set("X-Sourcegraph-Verbatim", "true")
}

func serveAssets(w http.ResponseWriter, r *http.Request) error {
	passThrough(w)
	f, err := assets.Assets.Open("/" + mux.Vars(r)["assetFile"])
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(b)
	return nil
}

func handler(f func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

var router *mux.Router

func init() {
	router = mux.NewRouter()
	router.StrictSlash(true)

	post := router.Methods("POST").Subrouter()
	post.Handle("/new", handler(serveNewIssue))
	post.Handle("/{issueNum:[0-9]+}/{replyNum:[0-9]+}/edit", handler(serveEditReply))
	post.Handle("/{issueNum:[0-9]+}/close", handler(serveCloseIssue))
	post.Handle("/{issueNum:[0-9]+}", handler(serveIssueReplyCreate))

	router.Handle("/", handler(serveMainPage))
	router.Handle("/assets/{assetFile:.*}", handler(serveAssets))
	router.Handle("/{issueNum:[0-9]+}", handler(serveIssue))
	router.Handle("/{issueNum:[0-9]+}/{replyNum:[0-9]+}/raw", handler(serveRawReply))
}

type Handler struct{}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}
	router.ServeHTTP(w, r)
}
