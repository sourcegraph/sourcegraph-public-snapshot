package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/peterhellberg/link"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Visibility string

const (
	Public   Visibility = "public"
	Private  Visibility = "private"
	Internal Visibility = "internal"
)

// Project is a GitLab project (equivalent to a GitHub repository).
type Project struct {
	ProjectCommon
	Visibility        Visibility     `json:"visibility"`                    // "private", "internal", or "public"
	ForkedFromProject *ProjectCommon `json:"forked_from_project,omitempty"` // If non-nil, the project from which this project was forked
	Archived          bool           `json:"archived"`
	StarCount         int            `json:"star_count"`
	ForksCount        int            `json:"forks_count"`
	EmptyRepo         bool           `json:"empty_repo"`
	DefaultBranch     string         `json:"default_branch"`
	Topics            []string       `json:"topics"`
}

type ProjectCommon struct {
	ID                int    `json:"id"`                  // ID of project
	PathWithNamespace string `json:"path_with_namespace"` // full path name of project ("namespace1/namespace2/name")
	Description       string `json:"description"`         // description of project
	WebURL            string `json:"web_url"`             // the web URL of this project ("https://gitlab.com/foo/bar")i
	HTTPURLToRepo     string `json:"http_url_to_repo"`    // HTTP clone URL
	SSHURLToRepo      string `json:"ssh_url_to_repo"`     // SSH clone URL ("git@example.com:foo/bar.git")
}

// Name returns the project name.
func (pc *ProjectCommon) Name() (string, error) {
	// Although there is a name field available in GitLab projects returned by
	// the REST API, we can't rely on it being in local caches because we haven't
	// previously requested it. Fortunately, we can figure it out from the
	// PathWithNamespace.
	parts := strings.Split(pc.PathWithNamespace, "/")
	if len(parts) < 2 {
		return "", errors.New("path with namespace does not include any namespaces")
	}

	return parts[len(parts)-1], nil
}

// Namespace returns the project namespace(s) as a slash separated string.
func (pc *ProjectCommon) Namespace() (string, error) {
	parts := strings.Split(pc.PathWithNamespace, "/")
	if len(parts) < 2 {
		return "", errors.New("path with namespace does not include any namespaces")
	}

	return strings.Join(parts[0:len(parts)-1], "/"), nil
}

// RequiresAuthentication reports whether this project requires authentication to view (i.e., its visibility is
// "private" or "internal").
func (p Project) RequiresAuthentication() bool {
	return p.Visibility == "private" || p.Visibility == "internal"
}

// ContentsVisible reports whether or not the repository contents of this project is visible to the user.
// Repo content visibility is determined by checking whether or not the default branch of the project
// was returned in the JSON response. If no default branch is returned it means that either the
// project has no repository initialised, or the user cannot see the contents of the repository.
func (p *Project) ContentsVisible() bool {
	return p.DefaultBranch != ""
}

func idCacheKey(id int) string                                  { return "1:" + strconv.Itoa(id) }
func pathWithNamespaceCacheKey(pathWithNamespace string) string { return "1:" + pathWithNamespace }

// MockGetProject_Return is called by tests to mock (*Client).GetProject.
func MockGetProject_Return(returns *Project) {
	MockGetProject = func(*Client, context.Context, GetProjectOp) (*Project, error) {
		return returns, nil
	}
}

type GetProjectOp struct {
	ID                int
	PathWithNamespace string
	CommonOp
}

// GetProject gets a project from GitLab by either ID or path with namespace.
func (c *Client) GetProject(ctx context.Context, op GetProjectOp) (*Project, error) {
	if op.ID != 0 && op.PathWithNamespace != "" {
		panic("invalid args (specify exactly one of id and pathWithNamespace)")
	}

	if MockGetProject != nil {
		return MockGetProject(c, ctx, op)
	}

	var key string
	if op.ID != 0 {
		key = idCacheKey(op.ID)
	} else {
		key = pathWithNamespaceCacheKey(op.PathWithNamespace)
	}
	return c.cachedGetProject(ctx, key, op.NoCache, func(ctx context.Context) (proj *Project, keys []string, err error) {
		keys = append(keys, key)
		proj, err = c.getProjectFromAPI(ctx, op.ID, op.PathWithNamespace)
		if proj != nil {
			// Add the cache key for the other kind of specifier (ID vs. path with namespace) so it's addressable by
			// both in the cache.
			if op.ID != 0 {
				keys = append(keys, pathWithNamespaceCacheKey(proj.PathWithNamespace))
			} else {
				keys = append(keys, idCacheKey(proj.ID))
			}
		}
		return proj, keys, err
	})
}

// cachedGetProject caches the getProjectFromAPI call.
func (c *Client) cachedGetProject(ctx context.Context, key string, forceFetch bool, getProjectFromAPI func(context.Context) (proj *Project, keys []string, err error)) (*Project, error) {
	if !forceFetch {
		if cached := c.getProjectFromCache(ctx, key); cached != nil {
			projectsGitLabCacheCounter.WithLabelValues("hit").Inc()
			if cached.NotFound {
				return nil, ErrProjectNotFound
			}
			return &cached.Project, nil
		}
	}

	proj, keys, err := getProjectFromAPI(ctx)
	if IsNotFound(err) {
		// Before we do anything, ensure we cache NotFound responses.
		// Do this if client is unauthed or authed, it's okay since we're only caching not found responses here.
		c.addProjectToCache(keys, &cachedProj{NotFound: true})
		projectsGitLabCacheCounter.WithLabelValues("notfound").Inc()
	}
	if err != nil {
		projectsGitLabCacheCounter.WithLabelValues("error").Inc()
		return nil, err
	}

	c.addProjectToCache(keys, &cachedProj{Project: *proj})
	projectsGitLabCacheCounter.WithLabelValues("miss").Inc()

	return proj, nil
}

var projectsGitLabCacheCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_projs_gitlab_cache_hit",
	Help: "Counts cache hits and misses for GitLab project metadata.",
}, []string{"type"})

type cachedProj struct {
	Project

	// NotFound indicates that the GitLab API reported that the project was not found.
	NotFound bool
}

// getProjectFromCache attempts to get a response from the redis cache.
// It returns nil error for cache-hit condition and non-nil error for cache-miss.
func (c *Client) getProjectFromCache(_ context.Context, key string) *cachedProj {
	b, ok := c.projCache.Get(strings.ToLower(key))
	if !ok {
		return nil
	}

	var cached cachedProj
	if err := json.Unmarshal(b, &cached); err != nil {
		return nil
	}

	return &cached
}

// addProjectToCache will cache the value for proj. The caller can provide multiple cache keys for the multiple
// ways that this project can be retrieved (e.g., both ID and path with namespace).
func (c *Client) addProjectToCache(keys []string, proj *cachedProj) {
	b, err := json.Marshal(proj)
	if err != nil {
		return
	}
	for _, key := range keys {
		c.projCache.Set(strings.ToLower(key), b)
	}
}

// getProjectFromAPI attempts to fetch a project from the GitLab API without use of the redis cache.
func (c *Client) getProjectFromAPI(ctx context.Context, id int, pathWithNamespace string) (proj *Project, err error) {
	var urlParam string
	if id != 0 {
		urlParam = strconv.Itoa(id)
	} else {
		urlParam = url.PathEscape(pathWithNamespace) // https://docs.gitlab.com/ce/api/README.html#namespaced-path-encoding
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("projects/%s", urlParam), nil)
	if err != nil {
		return nil, err
	}
	_, _, err = c.do(ctx, req, &proj)
	if IsNotFound(err) {
		err = &ProjectNotFoundError{Name: pathWithNamespace}
	}
	return proj, err
}

// ListProjects lists GitLab projects.
func (c *Client) ListProjects(ctx context.Context, urlStr string) (projs []*Project, nextPageURL *string, err error) {
	if MockListProjects != nil {
		return MockListProjects(c, ctx, urlStr)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}
	respHeader, _, err := c.do(ctx, req, &projs)
	if err != nil {
		return nil, nil, err
	}

	// Get URL to next page. See https://docs.gitlab.com/ee/api/README.html#pagination-link-header.
	if l := link.Parse(respHeader.Get("Link"))["next"]; l != nil {
		nextPageURL = &l.URI
	}

	// Add to cache.
	for _, proj := range projs {
		keys := []string{pathWithNamespaceCacheKey(proj.PathWithNamespace), idCacheKey(proj.ID)} // cache under multiple
		c.addProjectToCache(keys, &cachedProj{Project: *proj})
	}

	return projs, nextPageURL, nil
}

// Fork forks a GitLab project. If namespace is nil, then the project will be
// forked into the current user's namespace.
//
// If the project has already been forked, then the forked project is retrieved
// and returned.
func (c *Client) ForkProject(ctx context.Context, project *Project, namespace *string, name string) (*Project, error) {
	// Let's be optimistic and see if there's a fork already first, thereby
	// saving us an API call or two on the happy path.
	resolved, err := c.resolveNamespace(ctx, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "resolving namespace")
	}

	fork, err := c.getForkedProject(ctx, resolved, name)
	if err != nil {
		// An error that _isn't_ a not found error needs to be reported.
		if !IsNotFound(err) {
			return nil, errors.Wrap(err, "checking for previously forked project")
		}
	} else if err == nil {
		// Happy path: let's just return the fork, and we're done.
		return fork, nil
	}

	// Now we know we have to fork the project into the namespace.
	payload := struct {
		NamespacePath *string `json:"namespace_path,omitempty"`
		Name          string  `json:"name,omitempty"`
		// a path must be specified here otherwise it will use the original repo path, regardless of the new repo name
		Path *string `json:"path,omitempty"`
	}{
		NamespacePath: namespace,
		Name:          name,
		Path:          &name,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling payload")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("projects/%d/fork", project.ID), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	_, code, err := c.do(ctx, req, &fork)
	if code == http.StatusConflict {
		// 409 Conflict is returned if the fork already exists. While we should
		// have detected that earlier, it's possible — if unlikely — that someone
		// forked the project between the calls, so let's just roll with it. In
		// this case, we want to ignore the error generated by doWithBaseURL, and
		// instead get the forked project and return that.
		return c.getForkedProject(ctx, resolved, name)
	} else if err != nil {
		return nil, errors.Wrap(err, "forking project")
	}

	return fork, nil
}

func (c *Client) getForkedProject(ctx context.Context, namespace string, name string) (*Project, error) {
	// Note that we disable the cache when retrieving forked projects as it
	// interferes with the not found error detection in ForkProject.
	return c.GetProject(ctx, GetProjectOp{
		PathWithNamespace: namespace + "/" + name,
		CommonOp:          CommonOp{NoCache: true},
	})
}

func (c *Client) resolveNamespace(ctx context.Context, namespace *string) (string, error) {
	if namespace != nil {
		return *namespace, nil
	}

	user, err := c.GetUser(ctx, "")
	if err != nil {
		return "", err
	}

	return user.Username, nil
}
