package vcsclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/go-querystring/query"
	muxpkg "github.com/sourcegraph/mux"
	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

var ErrRepoNotExist = errors.New("repository does not exist on remote server")

func IsRepoNotExist(err error) bool {
	if err == nil {
		return false
	}
	if err == ErrRepoNotExist {
		return true
	}
	if err, ok := err.(*ErrorResponse); ok {
		return err.Message == ErrRepoNotExist.Error()
	}
	return err.Error() == ErrRepoNotExist.Error()
}

type repository struct {
	client   *Client
	repoPath string
}

var _ vcs.Repository = (*repository)(nil)

type RepositoryCloneUpdater interface {
	// CloneOrUpdate instructs the server to clone the repository so
	// it is available to the client via the API if it doesn't yet
	// exist, or update it from its default remote. The call blocks
	// until cloning finishes or fails.
	CloneOrUpdate(cloneInfo *CloneInfo) error
}

// CloneInfo is the information needed to clone a repository.
type CloneInfo struct {
	// VCS is the type of VCS (e.g., "git")
	VCS string

	// CloneURL is the remote URL from which to clone.
	CloneURL string

	// Additional options
	vcs.RemoteOpts
}

func (r *repository) CloneOrUpdate(cloneInfo *CloneInfo) error {
	url, err := r.url(RouteRepo, nil, nil)
	if err != nil {
		return err
	}

	req, err := r.client.NewRequest("POST", url.String(), cloneInfo)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req, nil)
	if err != nil {
		return err
	}
	if c := resp.StatusCode; c != http.StatusOK && c != http.StatusCreated {
		return fmt.Errorf("CloneOrUpdate: HTTP error %d", c)
	}

	return nil
}

func (r *repository) ResolveBranch(name string) (vcs.CommitID, error) {
	url, err := r.url(RouteRepoBranch, map[string]string{"Branch": name}, nil)
	if err != nil {
		return "", err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := r.client.doIgnoringRedirects(req)
	if err != nil {
		return "", err
	}

	return r.parseCommitIDInURL(resp.Header.Get("location"))
}

func (r *repository) ResolveRevision(spec string) (vcs.CommitID, error) {
	url, err := r.url(RouteRepoRevision, map[string]string{"RevSpec": spec}, nil)
	if err != nil {
		return "", err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := r.client.doIgnoringRedirects(req)
	if err != nil {
		return "", err
	}

	return r.parseCommitIDInURL(resp.Header.Get("location"))
}

func (r *repository) ResolveTag(name string) (vcs.CommitID, error) {
	url, err := r.url(RouteRepoTag, map[string]string{"Tag": name}, nil)
	if err != nil {
		return "", err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := r.client.doIgnoringRedirects(req)
	if err != nil {
		return "", err
	}

	return r.parseCommitIDInURL(resp.Header.Get("location"))
}

func (r *repository) parseCommitIDInURL(urlStr string) (vcs.CommitID, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	var info muxpkg.RouteMatch
	match := (*muxpkg.Router)(router).Match(&http.Request{Method: "GET", URL: url}, &info)
	if !match || info.Vars["CommitID"] == "" {
		return "", errors.New("failed to determine CommitID from URL")
	}

	return vcs.CommitID(info.Vars["CommitID"]), nil
}

func (r *repository) Branches(opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	url, err := r.url(RouteRepoBranches, nil, opt)
	if err != nil {
		return nil, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var branches []*vcs.Branch
	_, err = r.client.Do(req, &branches)
	if err != nil {
		return nil, err
	}

	return branches, nil
}

func (r *repository) Tags() ([]*vcs.Tag, error) {
	url, err := r.url(RouteRepoTags, nil, nil)
	if err != nil {
		return nil, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var tags []*vcs.Tag
	_, err = r.client.Do(req, &tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *repository) GetCommit(id vcs.CommitID) (*vcs.Commit, error) {
	url, err := r.url(RouteRepoCommit, map[string]string{"CommitID": string(id)}, nil)
	if err != nil {
		return nil, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var commit *vcs.Commit
	_, err = r.client.Do(req, &commit)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

// TotalCommitsHeader is the name of the HTTP header that contains the
// total number of commits in a call to Commits.
const TotalCommitsHeader = "x-vcsstore-total-commits"

func (r *repository) Commits(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	url, err := r.url(RouteRepoCommits, nil, opt)
	if err != nil {
		return nil, 0, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, 0, err
	}

	var commits []*vcs.Commit
	resp, err := r.client.Do(req, &commits)
	if err != nil {
		return nil, 0, err
	}

	total, err := strconv.ParseUint(string(resp.Header.Get(TotalCommitsHeader)), 10, 64)
	if err != nil {
		return nil, 0, err
	}

	return commits, uint(total), nil
}

func (r *repository) Committers(opt vcs.CommittersOptions) ([]*vcs.Committer, error) {
	url, err := r.url(RouteRepoCommitters, nil, opt)
	if err != nil {
		return nil, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var committers []*vcs.Committer
	_, err = r.client.Do(req, &committers)
	if err != nil {
		return nil, err
	}

	return committers, nil
}

// FileSystem returns a vfs.FileSystem that accesses the repository tree. The
// returned interface also satisfies vcsclient.FileSystem, which has an
// additional Get method that is useful for fetching all information about an
// entry in the tree.
func (r *repository) FileSystem(at vcs.CommitID) (vfs.FileSystem, error) {
	return &repositoryFS{
		at:   at,
		repo: r,
	}, nil
}

// router used to generate URLs for the vcsstore API.
var router = NewRouter(nil)

// url generates the URL to the named vcsstore API endpoint, using the
// specified route variables and query options.
func (r *repository) url(routeName string, routeVars map[string]string, opt interface{}) (*url.URL, error) {
	route := (*muxpkg.Router)(router).Get(routeName)
	if route == nil {
		return nil, fmt.Errorf("no API route named %q", route)
	}

	routeVarsList := make([]string, 2*len(routeVars))
	i := 0
	for name, val := range routeVars {
		routeVarsList[i*2] = name
		routeVarsList[i*2+1] = val
		i++
	}
	routeVarsList = append(routeVarsList, "RepoPath", r.repoPath)
	url, err := route.URL(routeVarsList...)
	if err != nil {
		return nil, err
	}

	// make the route URL path relative to BaseURL by trimming the leading "/"
	url.Path = strings.TrimPrefix(url.Path, "/")

	if opt != nil {
		err = addOptions(url, opt)
		if err != nil {
			return nil, err
		}
	}

	return url, nil
}

// addOptions adds the parameters in opt as URL query parameters to u. opt
// must be a struct whose fields may contain "url" tags.
func addOptions(u *url.URL, opt interface{}) error {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}

	qs, err := query.Values(opt)
	if err != nil {
		return err
	}

	u.RawQuery = qs.Encode()
	return nil
}
