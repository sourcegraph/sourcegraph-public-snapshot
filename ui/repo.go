package ui

import (
	"encoding/json"
	"errors"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func serveRepoCreate(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)
	e := json.NewEncoder(w)

	opt := struct {
		RepoURI string
	}{}
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}
	if opt.RepoURI == "" {
		log15.Warn("No repository URI provided with repo create request")
		return errors.New("Must provide a repository name")
	}

	_, err = cl.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
		URI: opt.RepoURI,
		VCS: "git",
	})
	if err != nil {
		log15.Error("failed to create repo", "error", err)
		return err
	}

	repoList, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: sourcegraph.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}

	return e.Encode(repoList.Repos)
}
