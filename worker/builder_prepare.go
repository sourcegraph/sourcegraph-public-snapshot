package worker

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"runtime"
	"strings"

	droneexec "github.com/drone/drone-exec/exec"
	"github.com/drone/drone-plugin-go/plugin"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/yaml.v2"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// prepare prepares internal server configuration for the build axes
// and build nodes.
func (b *builder) prepare(ctx context.Context) error {
	b.opt = droneexec.Options{
		Cache:  true,
		Clone:  true,
		Build:  true,
		Deploy: true,
		Notify: true,
		Debug:  true,
	}

	yamlBytes, err := yaml.Marshal(b.config)
	if err != nil {
		return err
	}

	droneBuild := plugin.Build{
		Commit: b.build.CommitID,
		Branch: b.build.Branch,
	}
	switch {
	case b.build.Branch != "":
		droneBuild.Ref = "refs/heads/" + b.build.Branch
	case b.build.Tag != "":
		droneBuild.Ref = "refs/tags/" + b.build.Tag
	default:
		// We need to fetch all branches to find the commit (git
		// doesn't let you just fetch a single commit; see the docs on
		// Build.Branch in sourcegraph.proto for an explanation).
		droneBuild.Ref = "" // fetches all branches from the remote
	}

	b.payload = droneexec.Payload{
		Workspace: &plugin.Workspace{},
		Build:     &droneBuild,
		Job:       &plugin.Job{},
		System: &plugin.System{
			Plugins: []string{
				"plugins/*",
				// Whitelist our internal plugins.
				"srclib/*",
				"sourcegraph/*",
				"sourcegraph-test/*",
				"library/alpine:*",
			},
		},
		Yaml: string(yamlBytes),
	}

	cloneURL, cloneNetrc, err := b.getCloneURLAndAuthForBuild(ctx)
	if err != nil {
		return err
	}
	repoLink, err := droneRepoLink(b.repo.HTTPCloneURL)
	if err != nil {
		return err
	}
	b.payload.Repo = &plugin.Repo{
		FullName:  b.repo.URI,
		Clone:     cloneURL,
		Link:      repoLink,
		IsPrivate: true,
		IsTrusted: true,
	}
	if cloneNetrc != nil {
		b.payload.Netrc = append(b.payload.Netrc, cloneNetrc)
	}

	// Get the netrc entry we need for srclib-importing (and otherwise
	// contacting the Sourcegraph server from the container). These
	// may be the same credentials as the clone netrc credentials, but
	// that's not true in all cases (e.g., clone credentials could be
	// for GitHub).
	hostNetrc, err := b.getHostNetrcEntry(ctx)
	if err != nil {
		return err
	}
	b.payload.Netrc = append(b.payload.Netrc, hostNetrc)

	return nil
}

// getCloneURLAndAuthForBuild returns the information necessary to
// clone repo during the build. The cloneURL never contains
// credentials itself (e.g.,
// "http://user:password@host.com/repo.git"); the credentials are
// returned in the username and password fields. This means that the
// cloneURL is safe to display in the build log.
//
// TODO(native-ci): look up and use SSH keys as well (e.g., with
// MirroredRepoSSHKeys.Get), not just HTTP clone credentials.
func (b *builder) getCloneURLAndAuthForBuild(ctx context.Context) (cloneURL string, netrc *plugin.NetrcEntry, err error) {
	cloneURL = b.repo.HTTPCloneURL
	host, err := urlHostNoPort(cloneURL)
	if err != nil {
		return
	}

	switch {
	case b.repo.Mirror: // Mirror repos that live elsewhere.
		if b.repo.Private {
			// Fetch auth to external server (e.g., GitHub.com
			// credentials).
			//
			// TODO(sqs!native-ci): Does the ext.AuthStore only work
			// when the worker is running locally and has access to
			// the SGPATH (i.e., when it's not a remote worker)?
			authStore := ext.AuthStore{}
			cred, err := authStore.Get(ctx, host)
			if err != nil {
				return "", nil, fmt.Errorf("unable to fetch credentials for host %q: %v", host, err)
			}
			netrc = &plugin.NetrcEntry{
				Machine:  host,
				Login:    "x-oauth-basic",
				Password: cred.Token,
			}
		}

	case b.repo.Origin == "": // Repos hosted on this server.
		// NOTE: This assumes that if the if-condition below
		// holds, the repo's HTTPCloneURL is on the trusted server. If
		// it's ever possible for the HTTPCloneURL to be on a
		// different server but still have this if-condition hold,
		// then we could leak the user's credentials.
		netrc, err = b.getHostNetrcEntry(ctx)
		if err != nil {
			return
		}
		if len(netrc.Password) > 255 {
			// This should not occur anymore, but it is very
			// difficult to debug if it does, so log it
			// anyway.
			log15.Warn("warning: Long repository password is incompatible with git < 2.0. If you see git authentication errors, upgrade to git 2.0+.", "repo", b.repo.URI, "password length", len(netrc.Password))
		}
	}

	// Make the URL and credentials valid for the Docker container,
	// not the host. See the doc for containerAddrForHost for more
	// information.
	host, cloneURL, err = containerAddrForHost(cloneURL)
	if err != nil {
		return
	}
	if netrc != nil {
		netrc.Machine = host
	}

	return
}

// getHostNetrcEntry creates a netrc entry that authorizes access to
// the Sourcegraph server.
func (b *builder) getHostNetrcEntry(ctx context.Context) (*plugin.NetrcEntry, error) {
	token := cli.Credentials.GetAccessToken()
	if token == "" {
		return nil, errors.New("can't generate local netrc entry: token is empty")
	}
	host, _, err := containerAddrForHost(conf.AppURL(ctx).String())
	if err != nil {
		return nil, err
	}
	return &plugin.NetrcEntry{
		Machine:  host,
		Login:    "x-oauth-basic",
		Password: token,
	}, nil
}

// parseURLOrGitSSHURL parses a URL and handles Git SSH URLs like
// "user@host:path/to/repo.git".
//
// TODO(native-ci): Use https://github.com/whilp/git-urls, but ask the
// person to add a license.
func parseURLOrGitSSHURL(urlStr string) (*url.URL, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	if u.Host != "" {
		return u, nil
	}

	cleanPath := func(p string) string {
		p = path.Clean(p)
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		return p
	}

	// Special-case Git URLs of the form "host:/path".
	if u.Scheme != "" && u.Opaque == "" && u.Path != "" {
		u.Host = u.Scheme
		u.Scheme = "git+ssh"
		return u, nil
	}

	// Special-case Git URLs of the form "host:path".
	if u.Scheme != "" && u.Opaque != "" && u.Path == "" {
		u.Host = u.Scheme
		u.Path = cleanPath(u.Opaque)
		u.Scheme = "git+ssh"
		u.Opaque = ""
		return u, nil
	}

	// Special-case Git URLs of the form "user@host:path".
	parts := strings.SplitN(u.Path, ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil, &url.Error{Op: "Parse(Git)", URL: urlStr, Err: errors.New("no path (follows host then ':')")}
	}
	return url.Parse("git+ssh://" + parts[0] + cleanPath(parts[1]))
}

// urlHostNoPort returns the host in the given URL string, without the port.
func urlHostNoPort(urlStr string) (string, error) {
	u, err := parseURLOrGitSSHURL(urlStr)
	if err != nil {
		return "", err
	}
	if host, _, _ := net.SplitHostPort(u.Host); host != "" {
		// Remove ":port".
		u.Host = host
	}
	return u.Host, nil
}

// droneRepoLink determines the link to tell Drone.
//
// The link determines where the repo is checked out to:
// /drone/src/LINK.
//
// For Go packages, we want LINK to match the import path of the repo
// whenever possible, so we use a heuristic. If the host is localhost,
// an IP address, or has a port number, then the repo is probably a
// mirror (because it's unlikely a real repo would live on a host with
// a local or ugly URL), and we chop off the hostname.
//
// TODO(native-ci): Handle the case when cloneURL is a git URL like
// "user@host:path/to/repo.git".
func droneRepoLink(cloneURL string) (string, error) {
	u, err := parseURLOrGitSSHURL(cloneURL)
	if err != nil {
		return "", err
	}

	// Clean path.
	u.Path = path.Clean(u.Path)
	u.Path = strings.TrimPrefix(u.Path, "/")
	u.Path = strings.TrimSuffix(u.Path, ".git")

	// Has port, or is IP address.
	if u.Host == "localhost" || strings.Contains(u.Host, ":") || net.ParseIP(u.Host) != nil {
		return u.Path, nil
	}

	return path.Join(u.Host, u.Path), nil
}

// containerAddrForHost tries to solves the following problem: If we
// pass a URL like "http://localhost:3080/my/repo" to a Docker
// container, its "localhost" is not the same as the host's
// "localhost," and so that URL will not work as intended inside the
// container.
//
// In this case, the host URL and netrc hostname must be changed to
// use an address at which the container can reach the host. If we're
// running a Docker daemon on localhost, this is fairly easy/reliable:
// just use the IP address of the docker0 network interface. In some
// cases it may be impossible to solve; for example, if your
// Sourcegraph server is firewalled off from the Docker containers.
func containerAddrForHost(hostURL string) (hostname, containerURL string, err error) {
	origHost, err := urlHostNoPort(hostURL)
	if err != nil {
		return "", "", err
	}

	if origHost == "localhost" {
		u, err := parseURLOrGitSSHURL(hostURL)
		if err != nil {
			return "", "", err
		}
		u.Host = strings.Replace(u.Host, origHost, containerHostname, 1)
		containerURL = u.String()
		hostname, err = urlHostNoPort(containerURL)
		if err != nil {
			return "", "", err
		}
	} else {
		hostname = origHost
		u, err := parseURLOrGitSSHURL(hostURL)
		if err != nil {
			return "", "", err
		}
		containerURL = u.String()
	}
	return
}

// containerHostname is the IP address of the host, as viewed by
// Docker containers running on the host.
//
// TODO(native-ci): Un-hardcode this IP address; determine it by
// actually querying the docker0 network interface.
var containerHostname = func() string {
	if runtime.GOOS == "darwin" {
		return "192.168.99.1" // Docker machine's vboxnet0
	}
	return "172.17.42.1" // Linux's docker0
}()
