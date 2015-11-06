package issues

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/shurcooL/github_flavored_markdown"
	"github.com/shurcooL/go-goon"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/issues/common"
	"src.sourcegraph.com/apps/issues/issues"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
)

var t *template.Template

func loadTemplates() error {
	var err error
	t = template.New("").Funcs(template.FuncMap{
		"dump": func(v interface{}) string { return goon.Sdump(v) },
		"jsonfmt": func(v interface{}) (string, error) {
			b, err := json.MarshalIndent(v, "", "\t")
			return string(b), err
		},
		"reltime": humanize.Time,
		"gfm":     func(s string) template.HTML { return template.HTML(github_flavored_markdown.Markdown([]byte(s))) },
		"event":   func(e issues.Event) event { return event{e} },
	})
	t, err = vfstemplate.ParseGlob(Assets, t, "/assets/*.tmpl")
	return err
}

type state struct {
	BaseState
}

type BaseState struct {
	ctx  context.Context
	req  *http.Request
	vars map[string]string

	repoSpec issues.RepoSpec

	common.State
}

func baseState(req *http.Request) BaseState {
	ctx := putil.Context(req)
	repoRevSpec, _ := pctx.RepoRevSpec(ctx)
	return BaseState{
		ctx:  ctx,
		req:  req,
		vars: mux.Vars(req),

		repoSpec: issues.RepoSpec{URI: repoRevSpec.URI},

		State: common.State{
			BaseURI:   pctx.BaseURI(ctx),
			ReqPath:   req.URL.Path,
			CSRFToken: pctx.CSRFToken(ctx),
		},
	}
}

func (s state) Tab() (issues.State, error) {
	return tab(s.req.URL.Query())
}

func (s state) Tabs() (template.HTML, error) {
	return tabs(&s, s.BaseURI+s.ReqPath, s.req.URL.RawQuery)
}

func (s state) Issues() ([]issues.Issue, error) {
	var opt issues.IssueListOptions
	switch selectedTab := s.req.URL.Query().Get(queryKeyState); selectedTab {
	case "": // Default. TODO: Make this cleaner.
		opt.State = issues.OpenState
	case string(issues.ClosedState):
		opt.State = issues.ClosedState
	}
	return is.List(s.ctx, s.repoSpec, opt)
}

func (s state) OpenCount() (uint64, error) {
	return is.Count(s.ctx, s.repoSpec, issues.IssueListOptions{State: issues.OpenState})
}

func (s state) ClosedCount() (uint64, error) {
	return is.Count(s.ctx, s.repoSpec, issues.IssueListOptions{State: issues.ClosedState})
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func (s state) Issue() (interface{}, error) {
	return is.Get(s.ctx, s.repoSpec, uint64(mustAtoi(s.vars["id"])))
}

func (s state) Comments() (interface{}, error) {
	return is.ListComments(s.ctx, s.repoSpec, uint64(mustAtoi(s.vars["id"])), nil)
}

func (s state) Events() (interface{}, error) {
	return is.ListEvents(s.ctx, s.repoSpec, uint64(mustAtoi(s.vars["id"])), nil)
}

func (s state) Items() (interface{}, error) {
	cs, err := is.ListComments(s.ctx, s.repoSpec, uint64(mustAtoi(s.vars["id"])), nil)
	if err != nil {
		return nil, err
	}
	es, err := is.ListEvents(s.ctx, s.repoSpec, uint64(mustAtoi(s.vars["id"])), nil)
	if err != nil {
		return nil, err
	}
	var items []issueItem
	for _, comment := range cs {
		items = append(items, issueItem{comment})
	}
	for _, event := range es {
		items = append(items, issueItem{event})
	}
	sort.Sort(byCreatedAt(items))
	return items, nil
}

var is issues.Service

// Apparently needed for "new-comment" component, etc.
func (state) CurrentUser() issues.User {
	return is.CurrentUser()
}

func mainHandler(w http.ResponseWriter, req *http.Request) {
	if err := loadTemplates(); err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := path.Base(req.URL.Path)
	state := state{
		BaseState: baseState(req),
	}
	err := t.ExecuteTemplate(w, tmpl+".tmpl", &state)
	if err != nil {
		log.Println("t.ExecuteTemplate:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func issuesHandler(w http.ResponseWriter, req *http.Request) {
	if err := loadTemplates(); err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := state{
		BaseState: baseState(req),
	}
	err := t.ExecuteTemplate(w, "issues.html.tmpl", &state)
	if err != nil {
		log.Println("t.ExecuteTemplate:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func issueHandler(w http.ResponseWriter, req *http.Request) {
	if err := loadTemplates(); err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := state{
		BaseState: baseState(req),
	}
	err := t.ExecuteTemplate(w, "issue.html.tmpl", &state)
	if err != nil {
		log.Println("t.ExecuteTemplate:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func debugHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Sourcegraph-Verbatim", "true")

	fmt.Println("debugHandler:", req.URL.Path)

	ctx := putil.Context(req)
	if repoRevSpec, ok := pctx.RepoRevSpec(ctx); ok {
		goon.DumpExpr(issues.RepoSpec{URI: repoRevSpec.URI})
	}
	goon.DumpExpr(pctx.RepoRevSpec(ctx))

	//io.WriteString(w, req.PostForm.Get("value"))
}

func createIssueHandler(w http.ResponseWriter, req *http.Request) {
	if err := loadTemplates(); err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := state{
		BaseState: baseState(req),
	}
	err := t.ExecuteTemplate(w, "new-issue.html.tmpl", &state)
	if err != nil {
		log.Println("t.ExecuteTemplate:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postCreateIssueHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Sourcegraph-Verbatim", "true")

	if err := req.ParseForm(); err != nil {
		log.Println("req.ParseForm:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := putil.Context(req)
	baseURI := pctx.BaseURI(ctx)
	repoRevSpec, _ := pctx.RepoRevSpec(ctx)

	issue := issues.Issue{
		Title: req.PostForm.Get("title"),
		Comment: issues.Comment{
			Body: req.PostForm.Get("body"),
		},
	}

	issue, err := is.Create(ctx, issues.RepoSpec{URI: repoRevSpec.URI}, issue)
	if err != nil {
		log.Println("is.Create:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s/issues/%d", baseURI, issue.ID)
}

func postEditIssueHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Sourcegraph-Verbatim", "true")

	if err := req.ParseForm(); err != nil {
		log.Println("req.ParseForm:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := putil.Context(req)
	vars := mux.Vars(req)
	repoRevSpec, _ := pctx.RepoRevSpec(ctx)

	var ir issues.IssueRequest
	err := json.Unmarshal([]byte(req.PostForm.Get("value")), &ir)
	if err != nil {
		log.Println("postEditIssueHandler json.Unmarshal:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	issue, err := is.Edit(ctx, issues.RepoSpec{URI: repoRevSpec.URI}, uint64(mustAtoi(vars["id"])), ir)
	if err != nil {
		log.Println("is.Edit:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Move to right place?
	issueEvent := issues.Event{
		Actor:     is.CurrentUser(),
		CreatedAt: time.Now(),
	}
	switch {
	case ir.State != nil && *ir.State == issues.OpenState:
		issueEvent.Type = issues.Reopened
	case ir.State != nil && *ir.State == issues.ClosedState:
		issueEvent.Type = issues.Closed
	case ir.Title != nil:
		issueEvent.Type = issues.Renamed
		issueEvent.Rename = &issues.Rename{
			From: "TODO",
			To:   *ir.Title,
		}
	}

	err = func(w io.Writer, issue issues.Issue) error {
		var resp = make(url.Values)

		var buf bytes.Buffer
		err := t.ExecuteTemplate(&buf, "issue-badge", issue.State)
		if err != nil {
			return err
		}
		resp.Set("issue-state-badge", buf.String())
		buf.Reset()
		err = t.ExecuteTemplate(&buf, "toggle-button", issue.State)
		if err != nil {
			return err
		}
		resp.Set("issue-toggle-button", buf.String())
		buf.Reset()
		err = t.ExecuteTemplate(&buf, "event", issueEvent)
		if err != nil {
			return err
		}
		resp.Set("new-event", buf.String())

		_, err = io.WriteString(w, resp.Encode())
		return err
	}(w, issue)
	if err != nil {
		log.Println("t.ExecuteTemplate:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postCommentHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Sourcegraph-Verbatim", "true")

	if err := req.ParseForm(); err != nil {
		log.Println("req.ParseForm:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := putil.Context(req)
	vars := mux.Vars(req)
	repoRevSpec, _ := pctx.RepoRevSpec(ctx)

	comment := issues.Comment{
		Body: req.PostForm.Get("value"),
	}

	comment, err := is.CreateComment(ctx, issues.RepoSpec{URI: repoRevSpec.URI}, uint64(mustAtoi(vars["id"])), comment)
	if err != nil {
		log.Println("is.CreateComment:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "comment", comment)
	if err != nil {
		log.Println("t.ExecuteTemplate:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
