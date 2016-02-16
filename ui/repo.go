package ui

import (
	"encoding/json"
	"errors"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sqs/pbtypes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/repoupdater"
	"src.sourcegraph.com/sourcegraph/util/eventsutil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoCreate(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
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

	_, err = apiclient.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
		URI: opt.RepoURI,
		VCS: "git",
	})
	if err != nil {
		log15.Error("failed to create repo", "error", err)
		return err
	}

	repoList, err := apiclient.Repos.List(ctx, &sourcegraph.RepoListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: sourcegraph.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}

	return e.Encode(repoList.Repos)
}

type repoInfo struct {
	URI     string
	Private bool
	Language string
}

func serveRepoMirror(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
	currentUser := handlerutil.UserFromRequest(r)
	if currentUser == nil {
		return errors.New("Must be authenticated to mirror repos")
	}
	e := json.NewEncoder(w)

	var data = struct {
		Repos []*repoInfo
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return err
	}

	var numPrivate, numPublic int32

	for _, repoInfo := range data.Repos {
		repoURI := repoInfo.URI

		// Perform the following operations locally (non-federated) because it's a private repo.
		_, err := apiclient.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
			URI:      repoURI,
			VCS:      "git",
			CloneURL: "https://" + repoURI + ".git",
			Mirror:   true,
			Private:  repoInfo.Private,
			Language: repoInfo.Language,
		})
		if grpc.Code(err) == codes.AlreadyExists {
			log15.Warn("repo already exists", "uri", repoURI)
		} else if err != nil {
			log15.Warn("user settings integration update failed", "uri", repoURI, "error", err)
			return err
		}

		if repoInfo.Private {
			numPrivate += 1
		} else {
			numPublic += 1
		}

		repoupdater.Enqueue(&sourcegraph.Repo{URI: repoURI})
	}

	eventsutil.LogAddMirrorRepos(ctx, numPrivate, numPublic)

	mirrorData, err := apiclient.MirrorRepos.GetUserData(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	return e.Encode(mirrorData)
}
