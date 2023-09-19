package images

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-enry/go-enry/v2/regex"
	"github.com/opencontainers/go-digest"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var listTagRoute = "https://%s/v2/%s/tags/list"
var fetchDigestRoute = "https://%s/v2/%s/manifests/%s"

// Registry abstracts interacting with various registries, such as
// GCR or Docker.io. There are subtle differences, mostly in how to authenticate.
type Registry interface {
	GetByTag(repo string, tag string) (*Repository, error)
	GetLatest(repo string) (*Repository, error)
	FetchDigest(repo string, tag string) (digest.Digest, error)
}

type Repository struct {
	registry string
	name     string
	org      string
	tag      string
	sha256   string
}

func (r *Repository) Ref() string {
	return fmt.Sprintf(
		"%s/%s/%s:%s@sha256:%s",
		r.registry,
		r.org,
		r.name,
		r.tag,
		r.sha256,
	)
}

func (r *Repository) Name() string {
	return r.name
}

var rawImageRegex = regex.MustCompile(`(?P<registry>[^\/]+)\/(?P<org>[\w-]+)\/(?P<repo>[\w-]+):(?P<tag>[\w.-]+)(?:@sha256:(?P<digest>\w+))?`)

func matchWithSubexps(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	res := map[string]string{}
	for i, name := range match {
		res[rawImageRegex.SubexpNames()[i]] = name
	}
	return res
}

// ParseRepository parses a raw image field into a struct. If we detect that
// the container is not a Sourcegraph service, it returns ErrNoUpdateNeeded.
func ParseRepository(rawImg string) (*Repository, error) {
	res := matchWithSubexps(rawImageRegex, rawImg)
	// If there's no registry set in the raw image string, it means
	// it's something like ubuntu:20.04, meaning it's not a Sourcegraph image.
	if res["registry"] == "" {
		return nil, ErrNoUpdateNeeded
	}

	// We also reference fully qualified images, but that are not
	// in our own registry, again, we can simply skip them.
	if !strings.Contains(res["org"], "sourcegraph") {
		return nil, ErrNoUpdateNeeded
	}

	// Finally, check if amongst the remaining images, if the one we're
	// looking at is a service that gets automatically updated.
	var found bool
	for _, app := range images.SourcegraphDockerImages {
		if res["repo"] == app {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrNoUpdateNeeded
	}

	return &Repository{
		registry: res["registry"],
		org:      res["org"],
		name:     res["repo"],
		tag:      res["tag"],
		sha256:   res["digest"],
	}, nil
}

// GCR provides access to Google Cloud Registry API.
type GCR struct {
	token string
	host  string
	org   string
}

// NewGCR creates a new GCR API client.
func NewGCR(host, org string) *GCR {
	return &GCR{
		org:  org,
		host: host,
	}
}

// LoadToken gets the access-token to reach GCR through the environment.
func (r *GCR) LoadToken() error {
	b, err := exec.Command("gcloud", "auth", "print-access-token").Output()
	if err != nil {
		return err
	}
	r.token = strings.TrimSpace(string(b))
	return nil
}

// FetchDigest returns the digest for a given container repository.
func (r *GCR) FetchDigest(repo string, tag string) (digest.Digest, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(fetchDigestRoute, r.host, r.org+"/"+repo, tag), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.token))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return "", errors.Newf("fetchDigest (%s) %s:%s, got %v: %s", r.host, repo, tag, resp.Status, string(data))
	}

	d := resp.Header.Get("Docker-Content-Digest")
	g, err := digest.Parse(d)
	if err != nil {
		return "", err
	}
	return g, nil

}

// GetByTag returns a container repository, on that registry, for a given service at
// a given tag.
func (r *GCR) GetByTag(repo string, tag string) (*Repository, error) {
	digest, err := r.FetchDigest(repo, tag)
	if err != nil {
		return nil, err
	}

	return &Repository{
		registry: r.host,
		name:     repo,
		org:      r.org,
		tag:      tag,
		sha256:   digest.String(),
	}, err
}

// GetLatest returns the latest container repository on that registry, according
// to the given predicate.
func (r *GCR) GetLatest(repo string) (*Repository, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(listTagRoute, r.host, r.org+"/"+repo), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return nil, errors.New(resp.Status + ": " + string(data))
	}
	result := struct {
		Tags []string
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	tag, err := findLatestMainTag(result.Tags)
	if err != nil {
		return nil, err
	}
	digest, err := r.FetchDigest(repo, tag)
	if err != nil {
		return nil, err
	}

	return &Repository{
		registry: r.host,
		name:     repo,
		org:      r.org,
		tag:      tag,
		sha256:   digest.String(),
	}, err
}
