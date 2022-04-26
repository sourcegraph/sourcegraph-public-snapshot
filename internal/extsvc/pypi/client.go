// Package pypi
//
// A wrapper around pypi's JSON API (https://warehouse.pypa.io/api-reference/json.html).
//
// Nomenclature:
//
// A "project" on PyPI is the name of a collection of releases and files, and
// information about them. Projects on PyPI are made and shared by other members
// of the Python community so that you can use them.
//
// A "release" on PyPI is a specific version of a project. For example, the
// requests project has many releases, like "requests 2.10" and "requests 1.2.1".
// A release consists of one or more "files".
//
// A "file", also known as a "package", on PyPI is something that you can
// download and install. Because of different hardware, operating systems, and
// file formats, a release may have several files (packages), like an archive
// containing source code or a binary
//
// https://pypi.org/help/#packages
//
package pypi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Client struct {
	// A list of pypi proxies.
	urls []string
	cli  httpcli.Doer

	// TODO: confirm which authentication pypi requires.
	credentials string

	// Self-imposed rate-limiter. Pypi.org does not impose a rate limiting policy.
	limiter *rate.Limiter

	// The name of this proxy. Used in error messages.
	proxy string
}

func NewClient(urn string, urls []string, cli httpcli.Doer) *Client {
	return &Client{
		urls:    urls,
		cli:     cli,
		limiter: ratelimit.DefaultRegistry.Get(urn),
	}
}

// ProjectInfo is a subset of the fields returned by Pypi's JSON API.
type ProjectInfo struct {
	Info Info `json:"info"`

	// URLs.
	URLS []ReleaseURL `json:"urls"`
}

type Info struct {
	Name string `json:"name"`

	// The version of the project.
	Version string `json:"version"`
}

type ReleaseURL struct {
	URL string `json:"url"`

	// TODO: maybe enum?
	// Either sdist, bdist_wheel, maybe bdist_egg?
	PackageType string `json:"packagetype"`
}

// Release returns metadata about a project at the specified version.
func (c *Client) Release(ctx context.Context, project string, version string) (*ProjectInfo, error) {
	// TODO: should the "pypi" path element be part of config? This might depend on how customer proxies are set up.
	data, err := c.get(ctx, "pypi", project, version, "json")
	if err != nil {
		return nil, err
	}
	pi := new(ProjectInfo)
	err = json.Unmarshal(data, pi)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// TODO: currently not used. Use this to get the latest version of a project.
// project returns metadata about an individual project at the LATEST version.
func (c *Client) project(ctx context.Context, project string) (*ProjectInfo, error) {
	data, err := c.get(ctx, "pypi", project, "json")
	if err != nil {
		return nil, err
	}
	pi := new(ProjectInfo)
	err = json.Unmarshal(data, pi)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// Unpacker abstracts the various archive formats pypi offers, such as tar.gzip and wheel.
type Unpacker interface {
	Unpack(dstDir string) error
}

func (c *Client) GetArchive(ctx context.Context, project string, version string) (Unpacker, error) {
	releaseInfo, err := c.Release(ctx, project, version)
	if err != nil {
		return nil, err
	}

	u, err := selectURL(releaseInfo)
	if err != nil {
		return nil, err
	}

	switch u.PackageType {

	// TODO: support wheels
	case "sdist":
		atype := toArchiveType(u.URL)
		if atype == "" {
			return nil, errors.Errorf("archive %s does not have a file extension for %s==%s", u.URL, project, version)
		}
		switch atype {
		case gztar:
			rc, err := c.fetchTarBall(ctx, u.URL)
			if err != nil {
				return nil, err
			}
			return &tarballUnpacker{rc}, nil
		default:
			return nil, errors.Errorf("unsupported archive type %s for %s==%s", atype, project, version)
		}
	default:
		return nil, errors.Errorf("unsupported package type %s for %s==%s", u.PackageType, project, version)
	}
}

type archiveType string

const (
	zip   archiveType = "zip"
	tar   archiveType = "tar"
	gztar archiveType = "gztar"
	xztar archiveType = "xztar"
	bztar archiveType = "bztar"
	ztar  archiveType = "ztar"
)

// https://docs.python.org/3/distutils/sourcedist.html
func toArchiveType(path string) archiveType {
	typeOrEmpty := func(t archiveType, path, ext string) archiveType {
		if filepath.Ext(path[:len(path)-len(ext)]) == ".tar" {
			return t
		}
		return ""

	}
	ext := filepath.Ext(path)
	switch ext {
	case ".zip":
		return zip
	case ".tar":
		return tar
	case ".gz":
		return typeOrEmpty(gztar, path, ext)
	case ".bz2":
		return typeOrEmpty(bztar, path, ext)
	case ".xz":
		return typeOrEmpty(xztar, path, ext)
	case ".Z":
		return typeOrEmpty(ztar, path, ext)
	}
	return ""
}

// Implements Unpacker.
type tarballUnpacker struct {
	rc io.ReadCloser
}

func (t tarballUnpacker) Unpack(dstDir string) error {
	return unpack.DecompressTgz(t.rc, dstDir)
}

func (c *Client) fetchTarBall(ctx context.Context, url string) (io.ReadCloser, error) {
	startWait := time.Now()
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	if d := time.Since(startWait); d > rateLimitingWaitThreshold {
		log15.Warn("%s proxy client self-enforced API rate limit: request delayed longer than expected due to rate limit", c.proxy, "delay", d)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	//TODO: copy-pasta from npm. Validate with pypi documentation.
	if c.credentials != "" {
		req.Header.Set("Authorization", "Bearer "+c.credentials)
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bodyBuffer bytes.Buffer
	if _, err := io.Copy(&bodyBuffer, resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, pypiError{resp.StatusCode, errors.New(bodyBuffer.String())}
	}

	return io.NopCloser(&bodyBuffer), nil
}

type pypiError struct {
	statusCode int
	err        error
}

func (n pypiError) Error() string {
	// TODO: copy-pasta from npm: confirm if this applies to pypi
	if 100 <= n.statusCode && n.statusCode <= 599 {
		return fmt.Sprintf("pypi HTTP response %d: %s", n.statusCode, n.err.Error())
	}
	return n.err.Error()
}

func selectURL(info *ProjectInfo) (ReleaseURL, error) {
	for _, u := range info.URLS {
		if u.PackageType == "sdist" {
			return u, nil
		}
	}

	// This release doesn't offer an archive. We will try to find a wheel that
	// targets platform "any" instead.
	for _, u := range info.URLS {
		if u.PackageType == "bdist_wheel" {
			p := getPlatform(u.URL)
			if p == "" {
				continue
			}
			if p == "any" {
				return u, nil
			}
		}
	}

	// Return the first wheel.
	for _, u := range info.URLS {
		if u.PackageType == "bdist_wheel" {
			return u, nil
		}
	}

	return ReleaseURL{}, errors.Errorf("%s==%s does not contain an source distribution or wheel", info.Info.Name, info.Info.Version)
}

// https://peps.python.org/pep-0491/#file-format
func getPlatform(wheel string) string {
	i := strings.LastIndexByte(wheel, '-')
	if i == -1 {
		return ""
	}
	ext := filepath.Ext(wheel)
	if ext != ".whl" {
		return ""
	}
	wheel = strings.TrimSuffix(wheel, ext)
	return wheel[i+1:]
}

// rateLimitingWaitThreshold is maximum rate limiting wait duration after which
// a warning log is produced to help site admins debug why syncing may be taking
// longer than expected.
const rateLimitingWaitThreshold = 200 * time.Millisecond

func (c *Client) get(ctx context.Context, project string, paths ...string) (respBody []byte, err error) {
	var (
		reqURL *url.URL
		req    *http.Request
	)

	for _, baseURL := range c.urls {
		startWait := time.Now()
		if err = c.limiter.Wait(ctx); err != nil {
			return nil, err
		}

		if d := time.Since(startWait); d > rateLimitingWaitThreshold {
			log15.Warn("%s proxy client self-enforced API rate limit: request delayed longer than expected due to rate limit", c.proxy, "delay", d)
		}

		reqURL, err = url.Parse(baseURL)
		if err != nil {
			return nil, errors.Errorf("invalid %s proxy URL %q", c.proxy, baseURL)
		}
		reqURL.Path = path.Join(project, path.Join(paths...))

		req, err = http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
		if err != nil {
			return nil, err
		}

		respBody, err = c.do(req)
		if err == nil || !errcode.IsNotFound(err) {
			break
		}
	}

	return respBody, err
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &Error{path: req.URL.Path, code: resp.StatusCode, message: string(bs), proxy: c.proxy}
	}

	return bs, nil
}

type Error struct {
	path    string
	code    int
	message string
	proxy   string
}

func (e *Error) Error() string {
	return fmt.Sprintf("bad %s proxy response with status code %d for %s: %s", e.proxy, e.code, e.path, e.message)
}
