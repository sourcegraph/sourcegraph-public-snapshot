package internal

import (
	"context"
	"math"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/search"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/syncx"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func (s *Server) searchWithObservability(ctx context.Context, tr trace.Trace, args *protocol.SearchRequest, onMatch func(*protocol.CommitMatch) error) (limitHit bool, err error) {
	searchStart := time.Now()

	searchRunning.Inc()
	defer searchRunning.Dec()

	tr.SetAttributes(
		args.Repo.Attr(),
		attribute.Bool("include_diff", args.IncludeDiff),
		attribute.String("query", args.Query.String()),
		attribute.Int("limit", args.Limit),
		attribute.Bool("include_modified_files", args.IncludeModifiedFiles),
	)
	defer func() {
		tr.AddEvent("done", attribute.Bool("limit_hit", limitHit))
		tr.SetError(err)
		searchDuration.
			WithLabelValues(strconv.FormatBool(err != nil)).
			Observe(time.Since(searchStart).Seconds())

		if honey.Enabled() || traceLogs {
			act := actor.FromContext(ctx)
			ev := honey.NewEvent("gitserver-search")
			ev.SetSampleRate(honeySampleRate("", act))
			ev.AddField("repo", args.Repo)
			ev.AddField("revisions", args.Revisions)
			ev.AddField("include_diff", args.IncludeDiff)
			ev.AddField("include_modified_files", args.IncludeModifiedFiles)
			ev.AddField("actor", act.UIDString())
			ev.AddField("query", args.Query.String())
			ev.AddField("limit", args.Limit)
			ev.AddField("duration_ms", time.Since(searchStart).Milliseconds())
			if err != nil {
				ev.AddField("error", err.Error())
			}
			if traceID := trace.ID(ctx); traceID != "" {
				ev.AddField("traceID", traceID)
				ev.AddField("trace", trace.URL(traceID, conf.DefaultClient()))
			}
			if honey.Enabled() {
				_ = ev.Send()
			}
			if traceLogs {
				s.Logger.Debug("TRACE gitserver search", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
			}
		}
	}()

	observeLatency := syncx.OnceFunc(func() {
		searchLatency.Observe(time.Since(searchStart).Seconds())
	})

	onMatchWithLatency := func(cm *protocol.CommitMatch) error {
		observeLatency()
		return onMatch(cm)
	}

	return s.search(ctx, args, onMatchWithLatency)
}

// search handles the core logic of the search. It is passed a matchesBuf so it doesn't need to
// concern itself with event types, and all instrumentation is handled in the calling function.
func (s *Server) search(ctx context.Context, args *protocol.SearchRequest, onMatch func(*protocol.CommitMatch) error) (limitHit bool, err error) {
	args.Repo = protocol.NormalizeRepo(args.Repo)
	if args.Limit == 0 {
		args.Limit = math.MaxInt32
	}

	// We used to have an `ensureRevision`/`CloneRepo` calls here that were
	// obsolete, because a search for an unknown revision of the repo (of an
	// uncloned repo) won't make it to gitserver and fail with an ErrNoResolvedRepos
	// and a related search alert before calling the gitserver.
	//
	// However, to protect for a weird case of getting an uncloned repo here (e.g.
	// via a direct API call), we leave a `repoCloned` check and return an error if
	// the repo is not cloned.
	dir := gitserverfs.RepoDirFromName(s.ReposDir, args.Repo)
	if !repoCloned(dir) {
		s.Logger.Debug("attempted to search for a not cloned repo")
		return false, &gitdomain.RepoNotExistError{
			Repo: args.Repo,
		}
	}

	mt, err := search.ToMatchTree(args.Query)
	if err != nil {
		return false, err
	}

	// Ensure that we populate ModifiedFiles when we have a DiffModifiesFile filter.
	// --name-status is not zero cost, so we don't do it on every search.
	hasDiffModifiesFile := false
	search.Visit(mt, func(mt search.MatchTree) {
		switch mt.(type) {
		case *search.DiffModifiesFile:
			hasDiffModifiesFile = true
		}
	})

	// Create a callback that detects whether we've hit a limit
	// and stops sending when we have.
	var sentCount atomic.Int64
	var hitLimit atomic.Bool
	limitedOnMatch := func(match *protocol.CommitMatch) {
		// Avoid sending if we've already hit the limit
		if int(sentCount.Load()) >= args.Limit {
			hitLimit.Store(true)
			return
		}

		sentCount.Add(int64(matchCount(match)))
		onMatch(match)
	}

	searcher := &search.CommitSearcher{
		Logger:               s.Logger,
		RepoName:             args.Repo,
		RepoDir:              dir.Path(),
		Revisions:            args.Revisions,
		Query:                mt,
		IncludeDiff:          args.IncludeDiff,
		IncludeModifiedFiles: args.IncludeModifiedFiles || hasDiffModifiesFile,
	}

	return hitLimit.Load(), searcher.Search(ctx, limitedOnMatch)
}

// matchCount returns either:
// 1) the number of diff matches if there are any
// 2) the number of messsage matches if there are any
// 3) one, to represent matching the commit, but nothing inside it
func matchCount(cm *protocol.CommitMatch) int {
	if len(cm.Diff.MatchedRanges) > 0 {
		return len(cm.Diff.MatchedRanges)
	}
	if len(cm.Message.MatchedRanges) > 0 {
		return len(cm.Message.MatchedRanges)
	}
	return 1
}
