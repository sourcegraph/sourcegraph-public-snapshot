package issues

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/shurcooL/go/gzip_file_server"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"

	//ghissues "src.sourcegraph.com/apps/issues/issues/github"
	fsissues "src.sourcegraph.com/apps/issues/issues/fs"
)

func Init() {
	err := loadTemplates()
	if err != nil {
		log.Fatalln("loadTemplates:", err)
	}

	//is = ghissues.NewService(nil)
	// TODO: Consider creating the issues2 dir when starting service, etc. It should exist when service is active.
	is = fsissues.NewService(filepath.Join(os.Getenv("SGPATH"), "issues2"))

	issues := http.NewServeMux()
	issues.HandleFunc("/pages/", mainHandler)
	r := mux.NewRouter()
	r.StrictSlash(true) // TODO: Make redirection work.
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// HACK: Redirect to canonical inner URL until I can get the top-level path working correctly.
		w.Header().Set("X-Sourcegraph-Verbatim", "true")
		http.Redirect(w, req, pctx.BaseURI(putil.Context(req))+"/issues", http.StatusSeeOther)
	}).Methods("GET")
	// /github.com/{owner}/{repo}/issues
	r.HandleFunc("/issues", issuesHandler).Methods("GET")
	r.HandleFunc("/issues/{id:[0-9]+}", issueHandler).Methods("GET")
	r.HandleFunc("/issues/{id:[0-9]+}/comment", postCommentHandler).Methods("POST")
	r.HandleFunc("/issues/{id:[0-9]+}/edit", postEditIssueHandler).Methods("POST")
	r.HandleFunc("/issues/new", createIssueHandler).Methods("GET")
	r.HandleFunc("/issues/new", postCreateIssueHandler).Methods("POST")
	issues.Handle("/", r)
	issues.Handle("/assets/", passThrough{gzip_file_server.New(Assets)})
	issues.HandleFunc("/debug", debugHandler)

	platform.RegisterFrame(platform.RepoFrame{
		ID:      "issues",
		Title:   "Issues",
		Icon:    "issue-opened",
		Handler: issues,
	})
}
