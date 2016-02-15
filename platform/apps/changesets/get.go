package changesets

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/rogpeppe/rog-go/parallel"

	"github.com/sourcegraph/mux"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sqs/pbtypes"
	notif "src.sourcegraph.com/apps/notifications/notifications"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/platform/notifications"

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
	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}
	repo, ok := pctx.RepoRevSpec(ctx)
	if !ok {
		return errors.New("no repo found in context")
	}
	var q struct {
		Closed   bool `schema:"closed"`
		Assigned bool `schema:"assigned"`
		Page     int  `schema:"page"`
	}
	schemaDecoder.Decode(&q, r.URL.Query())
	if q.Page == 0 {
		q.Page = 1
	}

	op := &sourcegraph.ChangesetListOp{
		Repo:   repo.URI,
		Closed: q.Closed,
		Open:   !q.Closed && !q.Assigned,
		ListOptions: sourcegraph.ListOptions{
			PerPage: 10,
			Page:    int32(q.Page),
		},
	}

	if q.Assigned {
		a := authpkg.ActorFromContext(ctx)
		op.NeedsReview = &sourcegraph.UserSpec{
			Login: a.Login,
			UID:   int32(a.UID),
		}
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
	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}
	v := mux.Vars(r)
	id, err := strconv.ParseInt(v["ID"], 10, 64)
	if err != nil {
		return err
	}
	rc, vc, err := GetRepoAndRevCommon(r)
	if err != nil {
		return err
	}

	// Parallel fetch from Changesets service
	var (
		par     = parallel.NewRun(3)
		cs      *sourcegraph.Changeset
		reviews *sourcegraph.ChangesetReviewList
		events  *sourcegraph.ChangesetEventList
		csErr   error
	)
	changesetSpec := sourcegraph.ChangesetSpec{
		Repo: vc.RepoRevSpec.RepoSpec,
		ID:   id,
	}
	reviewsSpec := &sourcegraph.ChangesetListReviewsOp{
		Repo:        vc.RepoRevSpec.RepoSpec,
		ChangesetID: id,
	}
	par.Do(func() error {
		cs, csErr = sg.Changesets.Get(ctx, &sourcegraph.ChangesetGetOp{
			Spec:              changesetSpec,
			FullReviewerUsers: true,
		})
		return csErr
	})
	par.Do(func() error {
		var err error
		reviews, err = sg.Changesets.ListReviews(ctx, reviewsSpec)
		return err
	})
	par.Do(func() error {
		var err error
		events, err = sg.Changesets.ListEvents(ctx, &changesetSpec)
		return err
	})
	err = par.Wait()
	if csErr != nil {
		err = csErr
	}
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("changeset does not exist")}
		}
		return err
	}

	// Fetch data which depends on the deltaspec concurrently
	var (
		ds      = cs.DeltaSpec
		delta   *sourcegraph.Delta
		baseTip *vcs.Commit
		files   *sourcegraph.DeltaFiles
	)
	par = parallel.NewRun(3)
	par.Do(func() error {
		// Compute delta (actual merge-base, commit IDs and build status for both revs)
		var err error
		delta, err = sg.Deltas.Get(ctx, &sourcegraph.DeltaSpec{
			Base: ds.Base,
			Head: ds.Head,
		})
		return err
	})
	par.Do(func() error {
		// Compute the tip of the base branch
		var err error
		baseTip, err = sg.Repos.GetCommit(ctx, &ds.Base)
		return err
	})
	par.Do(func() error {
		// Read and parse the .srcignore file.
		ignorePatterns, err := readSrcignore(ctx, vc.RepoRevSpec)
		if err != nil {
			return err
		}

		// If this route is for changes, request the diffs too
		files, err = sg.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{
			Ds: *ds,
			Opt: &sourcegraph.DeltaListFilesOptions{
				Filter: v["Filter"],
				Ignore: ignorePatterns,
			},
		})
		return err
	})
	err = par.Wait()
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
	// Augment commits with data from People
	augmentedCommits, err := handlerutil.AugmentCommits(ctx, delta.HeadRepo.URI, commitList.Commits)
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
		jiraURL, err := url.Parse(flags.JiraURL)
		if err != nil {
			return err
		}

		jiraIssues = make(map[string]string)
		ids := make([]string, 0)
		for _, commit := range commitList.Commits {
			ids = append(ids, parseJIRAIssues(commit.Message)...)
		}
		if cs.Description != "" {
			ids = append(ids, parseJIRAIssues(cs.Description)...)
		}
		for _, id := range ids {
			jiraIssues[id] = fmt.Sprintf("%s://%s/browse/%s", jiraURL.Scheme, jiraURL.Host, id)
		}
	}

	// Mark the changeset as read
	if notifications.Service != nil {
		notifications.Service.MarkRead(ctx, appID, notif.RepoSpec{URI: rc.Repo.URI}, uint64(id))
	}

	// If the source code contents of the diff are very large, then we
	// won't display file diffs, no point in sending them to the
	// client JSON-encoded. In fact, for large diffs this can mean the
	// difference between 317KB vs 16MB+ for the page!
	var overThreshold bool
	if st := files.DiffStat(); st.Added+st.Changed+st.Deleted > 5000 {
		overThreshold = true
		for _, fd := range files.FileDiffs {
			fd.FileDiffHunks = nil
			fd.FileDiff.Hunks = nil
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
		OverThreshold    bool // too large to display on one page
	}{
		RepoCommon:       *rc,
		RepoRevCommon:    *vc,
		FileFilter:       v["Filter"],
		ReviewGuidelines: guide,
		JiraIssues:       jiraIssues,
		OverThreshold:    overThreshold,

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

const defaultSrcignore = `
# Lines starting with a hash are ignored.

# Go
Godeps/*
*.gen.go
*.pb.go
*.pb_mock.go

`

const srcignore = ".srcignore"

// readSrcignore reads and parses the .srcignore file from the specified
// repository root, falling back to the global and default srcignore if needed.
//
// It returns the relevant '/'-separated glob patterns from the chosen file.
func readSrcignore(ctx context.Context, rev sourcegraph.RepoRevSpec) ([]string, error) {
	// Check the repository for a .srcignore file.
	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	te, err := sg.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: rev,
			Path:    srcignore,
		},
		Opt: &sourcegraph.RepoTreeGetOptions{
			ContentsAsString: true,
			GetFileOptions: sourcegraph.GetFileOptions{
				EntireFile: true,
			},
		},
	})
	if err == nil {
		return parseSrcignore(te.ContentsString), nil
	} else if grpc.Code(err) != codes.NotFound {
		return nil, err
	}

	// Check for a global one then.
	data, err := ioutil.ReadFile(filepath.Join(os.Getenv("SGPATH"), srcignore))
	if err == nil {
		return parseSrcignore(string(data)), nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Use a default .srcignore file.
	return parseSrcignore(defaultSrcignore), nil
}

// parseSrcignore parses a .srcignore file and returns the relevant unix glob
// patterns from the file.
func parseSrcignore(fileContents string) []string {
	// Accumulate pattern lines (ignoring empty lines + # comments).
	var patterns []string
	for _, line := range strings.Split(fileContents, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// byDate implements the sorting interface for sorting
// a list of commits by authorship date.
type byDate []*vcs.Commit

func (b byDate) Len() int { return len(b) }

func (b byDate) Less(i, j int) bool {
	return b[i].Author.Date.Time().Before(b[j].Author.Date.Time())
}

func (b byDate) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
