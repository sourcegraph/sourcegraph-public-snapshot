package router

import (
	"fmt"
	"net/url"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"

	"strconv"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func (r *Router) URLToBlogAtomFeed() *url.URL {
	return r.URLTo(BlogIndexAtom, "Format", ".atom")
}

func (r *Router) URLToBlogPost(slug string) *url.URL {
	return r.URLTo(BlogPost, "Slug", slug)
}

func (r *Router) URLToUser(userSpec string) *url.URL {
	return r.URLToUserSubroute(User, userSpec)
}

func (r *Router) URLToUserSubroute(routeName string, userSpec string) *url.URL {
	return r.URLTo(routeName, "User", userSpec)
}

func (r *Router) URLToRepo(uri string) *url.URL {
	return r.URLToRepoSubroute(Repo, uri)
}

func (r *Router) URLToRepoDiscussion(uri string, id int64) *url.URL {
	return r.URLTo(RepoDiscussion, "Repo", uri, "ID", strconv.FormatInt(id, 10))
}

func (r *Router) URLToRepoChangeset(uri string, id int64) *url.URL {
	return r.URLTo(Changeset, "Repo", uri, "ID", strconv.FormatInt(id, 10))
}

func (r *Router) URLToRepoChangesets(uri string, closed bool) *url.URL {
	url := r.URLTo(ChangesetList, "Repo", uri)
	if closed {
		q := url.Query()
		q.Set("Closed", "1")
		url.RawQuery = q.Encode()
	}
	return url
}

func (r *Router) URLToRepoRev(repoURI string, rev string) (*url.URL, error) {
	return r.URLToRepoSubrouteRev(Repo, repoURI, rev)
}

func (r *Router) URLToRepoBuild(repo string, commitID string, attempt uint32) *url.URL {
	return r.URLToRepoBuildSubroute(RepoBuild, repo, commitID, attempt)
}

func (r *Router) URLToRepoBuildSubroute(routeName string, repo string, commitID string, attempt uint32) *url.URL {
	return r.URLTo(routeName, "Repo", repo, "CommitID", commitID, "Attempt", strconv.Itoa(int(attempt)))
}

func (r *Router) URLToRepoBuildTaskSubroute(routeName string, repo string, commitID string, attempt uint32, taskID int64) *url.URL {
	return r.URLTo(routeName, "Repo", repo, "CommitID", commitID, "Attempt", strconv.Itoa(int(attempt)), "TaskID", strconv.Itoa(int(taskID)))
}

func (r *Router) URLToRepoSubroute(routeName string, uri string) *url.URL {
	return r.URLTo(routeName, "Repo", uri)
}

func (r *Router) URLToRepoSubrouteRev(routeName string, repoURI string, rev string) (*url.URL, error) {
	return r.URLToOrError(routeName, "Repo", repoURI, "Rev", rev)
}

func (r *Router) URLToRepoCompare(repoURI, rev, head string) (*url.URL, error) {
	return r.URLToOrError(RepoCompare, "Repo", repoURI, "Rev", rev, "Head", head)
}

func (r *Router) URLToRepoApp(repoURI string, appID string) (*url.URL, error) {
	return r.URLToOrError(RepoAppFrame, "Repo", repoURI, "App", appID, "AppPath", "")
}

func (r *Router) URLToRepoTreeEntry(repoURI string, rev interface{}, path string) *url.URL {
	return r.URLToRepoTreeEntrySubroute(RepoTree, repoURI, commitIDStr(rev), path)
}

func (r *Router) URLToRepoTreeEntrySubroute(routeName string, repo string, rev interface{}, path string) *url.URL {
	return r.URLTo(routeName, "Repo", repo, "Rev", commitIDStr(rev), "Path", path)
}

func (r *Router) URLToRepoTreeEntrySpec(e sourcegraph.TreeEntrySpec) *url.URL {
	return r.URLTo(RepoTree, "Repo", e.RepoRev.RepoSpec.SpecString(), "Rev", e.RepoRev.Rev, "CommitID", e.RepoRev.CommitID, "Path", e.Path)
}

func (r *Router) URLToRepoSearch(repoURI, rev, query string) (*url.URL, error) {
	var url *url.URL
	var err error
	if rev != "" {
		url, err = r.URLToRepoSubrouteRev(RepoSearch, repoURI, rev)
	} else {
		url = r.URLToRepoSubroute(RepoSearch, repoURI)
	}
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("q", query)
	url.RawQuery = q.Encode()
	return url, nil
}

func (r *Router) URLToSourceboxFile(e sourcegraph.TreeEntrySpec, format string) *url.URL {
	return r.URLTo(SourceboxFile, "Repo", e.RepoRev.RepoSpec.SpecString(), "Rev", e.RepoRev.Rev, "CommitID", e.RepoRev.CommitID, "Path", e.Path, "Format", format)
}

func (r *Router) URLToRepoTreeEntryLines(repoURI string, rev, path string, startLine int) *url.URL {
	u := r.URLTo(RepoTree, "Repo", repoURI, "Rev", rev, "Path", path)
	u.Fragment = fmt.Sprintf("L%d", startLine)
	return u
}

func (r *Router) URLToSearch(query string) *url.URL {
	url := r.URLTo(SearchResults)
	q := url.Query()
	q.Set("q", query)
	url.RawQuery = q.Encode()
	return url
}

func (r *Router) URLToDef(key graph.DefKey) *url.URL {
	return r.URLToDefSubroute(Def, key)
}

func (r *Router) URLToDefSubroute(routeName string, key graph.DefKey) *url.URL {
	return r.URLTo(routeName, "Repo", string(key.Repo), "UnitType", key.UnitType, "Unit", key.Unit, "Path", string(key.Path))
}

func (r *Router) URLToDefAtRev(key graph.DefKey, rev interface{}) *url.URL {
	return r.URLToDefAtRevSubroute(Def, key, rev)
}

func (r *Router) URLToDefAtRevSubroute(routeName string, key graph.DefKey, rev interface{}) *url.URL {
	return r.URLTo(routeName, "Repo", string(key.Repo), "Rev", commitIDStr(rev), "UnitType", key.UnitType, "Unit", key.Unit, "Path", string(key.Path))
}

func (r *Router) URLToSourceboxDef(key graph.DefKey, format string) *url.URL {
	return r.URLTo(SourceboxDef, "Repo", string(key.Repo), "UnitType", key.UnitType, "Unit", key.Unit, "Path", string(key.Path), "Format", format)
}

func (r *Router) URLToRepoCommit(repoURI string, commitID interface{}) *url.URL {
	return r.URLTo("repo.commit", "Repo", repoURI, "Rev", commitIDStr(commitID))
}

func commitIDStr(commitID interface{}) string {
	if v, ok := commitID.(vcs.CommitID); ok {
		return string(v)
	}
	return commitID.(string)
}
