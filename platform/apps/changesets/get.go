package changesets

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	"github.com/sourcegraph/mux"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

// serveList serves the page that displays the list of changesets on this
// repository.
func serveList(w http.ResponseWriter, r *http.Request) error {
	ctx := putil.Context(r)
	sg := sourcegraph.NewClientFromContext(ctx)
	repo, ok := pctx.RepoRevSpec(ctx)
	if !ok {
		return errors.New("no repo found in context")
	}
	var q struct {
		Closed bool `schema:"closed"`
		Page   int  `schema:"page"`
	}
	schemaDecoder.Decode(&q, r.URL.Query())
	if q.Page == 0 {
		q.Page = 1
	}

	op := &sourcegraph.ChangesetListOp{
		Repo:   repo.URI,
		Closed: q.Closed,
		Open:   !q.Closed,
		ListOptions: sourcegraph.ListOptions{
			PerPage: 10,
			Page:    int32(q.Page),
		},
	}
	list, err := sg.Changesets.List(ctx, op)
	if err != nil {
		return err
	}

	// TODO(slimsag): This is hacky. Our storage backend should tell us if we have
	// more.
	nextPageURL := ""
	prevPageURL := ""

	if op.ListOptions.Page > 1 {
		query := r.URL.Query()
		query.Set("page", strconv.FormatInt(int64(op.ListOptions.Page-1), 10))
		prevPageURL = "?" + query.Encode()
	}

	op.ListOptions.Page++
	nextList, err := sg.Changesets.List(ctx, op)
	if err == nil && len(nextList.Changesets) > 0 {
		query := r.URL.Query()
		query.Set("page", strconv.FormatInt(int64(op.ListOptions.Page), 10))
		nextPageURL = "?" + query.Encode()
	}

	return executeTemplate(w, r, "list.html", &struct {
		TmplCommon
		Repo                     sourcegraph.RepoRevSpec
		Op                       *sourcegraph.ChangesetListOp
		List                     []*sourcegraph.Changeset
		NextPageURL, PrevPageURL string
	}{
		Op:          op,
		Repo:        repo,
		List:        list.Changesets,
		NextPageURL: nextPageURL,
		PrevPageURL: prevPageURL,
	})
}

// serveChangeset serves the page that displays a changeset.
func serveChangeset(w http.ResponseWriter, r *http.Request) error {
	ctx := putil.Context(r)
	sg := sourcegraph.NewClientFromContext(ctx)
	v := mux.Vars(r)
	id, err := strconv.ParseInt(v["ID"], 10, 64)
	if err != nil {
		return err
	}
	rc, vc, err := GetRepoAndRevCommon(r)
	if err != nil {
		return err
	}
	// Retrieve Changeset
	changesetSpec := &sourcegraph.ChangesetSpec{
		Repo: vc.RepoRevSpec.RepoSpec,
		ID:   id,
	}
	cs, err := sg.Changesets.Get(ctx, changesetSpec)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("changeset does not exist")}
		}
		return err
	}
	ds := cs.DeltaSpec
	// Compute delta (actual merge-base, commit IDs and build status for both revs)
	delta, err := sg.Deltas.Get(ctx, &sourcegraph.DeltaSpec{
		Base: ds.Base,
		Head: ds.Head,
	})
	if err != nil {
		return err
	}
	reviews, err := sg.Changesets.ListReviews(ctx, &sourcegraph.ChangesetListReviewsOp{
		Repo:        vc.RepoRevSpec.RepoSpec,
		ChangesetID: cs.ID,
	})
	if err != nil {
		return err
	}
	// Compute the tip of the base branch
	baseTip, err := sg.Repos.GetCommit(ctx, &ds.Base)
	if err != nil {
		return err
	}
	// Retrieve commit list
	commitList, err := sg.Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: ds.Base.RepoSpec,
		Opt: &sourcegraph.RepoListCommitsOptions{
			Base:        string(delta.BaseCommit.ID),
			Head:        string(delta.HeadCommit.ID),
			ListOptions: sourcegraph.ListOptions{PerPage: -1},
		},
	})
	if err != nil {
		return err
	}
	sort.Sort(byDate(commitList.Commits))
	// If this route is for changes, request the diffs too
	var (
		files  *sourcegraph.DeltaFiles
		filter string
	)
	if f, ok := v["Filter"]; ok {
		filter = f
	}
	opt := sourcegraph.DeltaListFilesOptions{
		Formatted: false,
		Tokenized: true,
		Filter:    filter,
	}
	files, err = sg.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{
		Ds:  *ds,
		Opt: &opt,
	})
	if err != nil {
		return err
	}
	// Augment commits with data from People
	augmentedCommits := make([]*payloads.AugmentedCommit, len(commitList.Commits))
	for i, c := range commitList.Commits {
		augmentedCommits[i], err = handlerutil.AugmentCommit(r, delta.HeadRepo.URI, c)
		if err != nil {
			return err
		}
	}
	events, err := sg.Changesets.ListEvents(ctx, changesetSpec)
	if err != nil {
		return err
	}
	guide := pbtypes.HTML{}
	if flags.ReviewGuidelines != "" {
		if slurp, err := ioutil.ReadFile(flags.ReviewGuidelines); err == nil {
			if mdd, err := sg.Markdown.Render(ctx, &sourcegraph.MarkdownRenderOp{
				Markdown: slurp,
				Opt:      sourcegraph.MarkdownOpt{EnableCheckboxes: true},
			}); err == nil {
				guide.HTML = string(mdd.Rendered)
			}
		}
	}
	// Generate JIRA issue links
	var jiraIssues map[string]string
	if flags.JiraURL != "" {
		jiraIssues = make(map[string]string)
		ids := make([]string, 0)
		for _, commit := range commitList.Commits {
			ids = append(ids, parseJIRAIssues(commit.Message)...)
		}
		if cs.Description != "" {
			ids = append(ids, parseJIRAIssues(cs.Description)...)
		}
		protocol := "http"
		if flags.JiraTLS {
			protocol = "https"
		}
		for _, id := range ids {
			jiraIssues[id] = fmt.Sprintf("%s://%s/browse/%s", protocol, flags.JiraURL, id)
		}
	}

	return executeTemplate(w, r, "changeset.html", &struct {
		TmplCommon
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		payloads.Changeset

		FileFilter       string
		ReviewGuidelines pbtypes.HTML
		JiraIssues       map[string]string
	}{
		RepoCommon:       *rc,
		RepoRevCommon:    *vc,
		FileFilter:       filter,
		ReviewGuidelines: guide,
		JiraIssues:       jiraIssues,

		Changeset: payloads.Changeset{
			Changeset: cs,
			Delta:     delta,
			BaseTip:   baseTip,
			Commits:   augmentedCommits,
			Files:     files,
			Reviews:   reviews.Reviews,
			Events:    events.Events,
		},
	})
}

// byDate implements the sorting interface for sorting
// a list of commits by authorship date.
type byDate []*vcs.Commit

func (b byDate) Len() int { return len(b) }

func (b byDate) Less(i, j int) bool {
	return b[i].Author.Date.Time().Before(b[j].Author.Date.Time())
}

func (b byDate) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
