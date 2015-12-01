package handlerutil

import (
	"net/http"
	"sync"
	"time"

	"code.google.com/p/rog-go/parallel"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// AugmentCommits fills in fields on this package's commit type for
// presentation in the app.
func AugmentCommits(r *http.Request, repoURI string, commits []*vcs.Commit) ([]*payloads.AugmentedCommit, error) {
	ctx := httpctx.FromRequest(r)
	cl := APIClient(r)

	people := map[string]*sourcegraph.Person{}
	for _, commit0 := range commits {
		if commit0.Author.Email != "" {
			people[commit0.Author.Email] = nil
		}
		if commit0.Committer != nil && commit0.Committer.Email != commit0.Author.Email {
			people[commit0.Committer.Email] = nil
		}
	}

	peopleMu := sync.Mutex{}
	par := parallel.NewRun(4)
	for email := range people {
		email := email
		par.Do(func() error {
			author, err := cl.People.Get(ctx, &sourcegraph.PersonSpec{Email: email})
			if err != nil {
				return err
			}
			peopleMu.Lock()
			people[email] = author
			peopleMu.Unlock()
			return nil
		})
	}
	err := par.Wait()
	if err != nil {
		return nil, err
	}

	// We now have all the emails, lets construct the augmented commit list
	augmentedCommits := make([]*payloads.AugmentedCommit, len(commits))
	for i, commit0 := range commits {
		var author *sourcegraph.Person
		if commit0.Author.Email != "" {
			author = people[commit0.Author.Email]
		}
		var committer *sourcegraph.Person
		if commit0.Committer != nil && commit0.Committer.Email != commit0.Author.Email {
			committer = people[commit0.Committer.Email]
		}
		augmentedCommits[i] = &payloads.AugmentedCommit{
			Commit:          commit0,
			AuthorPerson:    author,
			CommitterPerson: committer,
			RepoURI:         repoURI,
		}
	}
	return augmentedCommits, nil
}

// DayOfAugmentedCommits is the same as DayOfCommits but its commits
// are augmentedCommits (which is the type we need to render a commit
// for the client).
type DayOfAugmentedCommits struct {
	Start   time.Time
	Commits []*payloads.AugmentedCommit
}

// AugmentAndGroupCommitsByDay runs both AugmentCommit and
// GroupCommitsByDay over the list of commits. It only performs the
// work of coalescing results from those two functions; all actual
// work is performed by them.
func AugmentAndGroupCommitsByDay(r *http.Request, commits []*vcs.Commit, repoURI string) ([]*DayOfAugmentedCommits, error) {
	days := GroupCommitsByDay(commits)
	augDays := make([]*DayOfAugmentedCommits, len(days))
	for i, day := range days {
		augDays[i] = &DayOfAugmentedCommits{
			Start:   day.Start,
			Commits: make([]*payloads.AugmentedCommit, len(day.Commits)),
		}
		augCommits, err := AugmentCommits(r, repoURI, day.Commits)
		if err != nil {
			return nil, err
		}
		for j := range day.Commits {
			augDays[i].Commits[j] = augCommits[j]
		}
	}
	return augDays, nil
}

// DayOfCommits represents a day's worth of commits, grouped by
// GroupCommitsByDay.
type DayOfCommits struct {
	Start   time.Time // start instant of day
	Commits []*vcs.Commit
}

// GroupCommitsByDay groups commits by day based on their commit
// date. This differs from the topological sort that `git log` shows,
// but it's how GitHub sorts commits, so it's probably OK.
//
// This function assumes that commits is already sorted by commit date
// (newest first).
//
// The exact handling of timezones is undefined right now. A
// reasonable assumption for end users of this function is that the
// times are all converted to UTC.
func GroupCommitsByDay(commits []*vcs.Commit) []*DayOfCommits {
	var days []*DayOfCommits
	for _, c := range commits {
		var d time.Time
		if c.Committer != nil {
			d = c.Committer.Date.Time()
		} else {
			// TODO(sqs): for hg commits (which don't have a committer
			// date) or git commits without a committer date, how to
			// handle them? preserve topo-sort?
			d = c.Author.Date.Time()
		}
		d = d.In(time.UTC)
		if len(days) == 0 || days[len(days)-1].Start.After(d) {
			// We've reached the previous day (or need to make the
			// first day).
			year, month, day := d.Date()
			days = append(days, &DayOfCommits{Start: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)})
		}
		day := days[len(days)-1]
		day.Commits = append(day.Commits, c)
	}
	return days
}
