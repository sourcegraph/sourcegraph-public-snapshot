package worker

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"runtime"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	droneexec "github.com/drone/drone-exec/exec"
	"github.com/drone/drone-plugin-go/plugin"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	httpapirouter "src.sourcegraph.com/sourcegraph/httpapi/router"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
	"src.sourcegraph.com/sourcegraph/sgx/client"
	"src.sourcegraph.com/sourcegraph/worker/builder"
)

func configureBuild(ctx context.Context, build *sourcegraph.Build) (*builder.Builder, error) {
	var b builder.Builder

	cl := sourcegraph.NewClientFromContext(ctx)

	repoSpec := sourcegraph.RepoSpec{URI: build.Repo}
	repoRev := sourcegraph.RepoRevSpec{
		RepoSpec: repoSpec,
		Rev:      build.CommitID,
		CommitID: build.CommitID,
	}
	repo, err := cl.Repos.Get(ctx, &repoRev.RepoSpec)
	if err != nil {
		return nil, err
	}

	// Read existing .drone.yml file.
	file, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{RepoRev: repoRev, Path: ".drone.yml"},
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return nil, err
	} else if err == nil {
		b.Payload.Yaml = string(file.Contents)
		b.DroneYMLFileExists = true
	}

	// Drone build
	b.Payload.Build = &plugin.Build{
		Commit: build.CommitID,
		Branch: build.Branch,
	}
	switch {
	case build.Branch != "":
		b.Payload.Build.Ref = "refs/heads/" + build.Branch
	case build.Tag != "":
		b.Payload.Build.Ref = "refs/tags/" + build.Tag
	default:
		// We need to fetch all branches to find the commit (git
		// doesn't let you just fetch a single commit; see the docs on
		// Build.Branch in sourcegraph.proto for an explanation).
		b.Payload.Build.Ref = "" // fetches all branches from the remote
	}

	// Drone repo
	cloneURL, cloneNetrc, err := getCloneURLAndAuthForBuild(ctx, repo)
	if err != nil {
		return nil, err
	}
	repoLink, err := droneRepoLink(repo.HTTPCloneURL)
	if err != nil {
		return nil, err
	}
	b.Payload.Repo = &plugin.Repo{
		FullName:  build.Repo,
		Clone:     cloneURL,
		Link:      repoLink,
		IsPrivate: true,
		IsTrusted: true,
	}
	if cloneNetrc != nil {
		b.Payload.Netrc = append(b.Payload.Netrc, cloneNetrc)
	}

	// Get the netrc entry we need for srclib-importing (and otherwise
	// contacting the Sourcegraph server from the container). These
	// may be the same credentials as the clone netrc credentials, but
	// that's not true in all cases (e.g., clone credentials could be
	// for GitHub).
	hostNetrc, err := getHostNetrcEntry(ctx)
	if err != nil {
		return nil, err
	}
	b.Payload.Netrc = append(b.Payload.Netrc, hostNetrc)

	// Drone other payload settings
	b.Payload.Workspace = &plugin.Workspace{}
	b.Payload.Job = &plugin.Job{}
	b.Payload.System = &plugin.System{
		Plugins: []string{
			"plugins/*",
			// Whitelist our internal plugins.
			"srclib/*",
			"sourcegraph/*",
			"sourcegraph-test/*",
			"library/alpine:*",
		},
	}

	// Drone options
	//
	// TODO(sqs): add these fields to the Sourcegraph build's
	// BuildConfig and use those values instead of always setting
	// true.
	b.Options = droneexec.Options{
		Cache:  true,
		Clone:  true,
		Build:  true,
		Deploy: false,
		Notify: true,
		Debug:  true,
	}

	// SrclibImportURL
	srclibImportURL, err := getSrclibImportURL(ctx, repoRev)
	if err != nil {
		return nil, err
	}
	b.SrclibImportURL = srclibImportURL

	// Inventory
	b.Inventory = func(ctx context.Context) (*inventory.Inventory, error) {
		return cl.Repos.GetInventory(ctx, &repoRev)
	}

	// CreateTasks
	b.CreateTasks = func(ctx context.Context, labels []string) ([]builder.TaskState, error) {
		tasks := make([]*sourcegraph.BuildTask, len(labels))
		for i, label := range labels {
			tasks[i] = &sourcegraph.BuildTask{Label: label}
		}
		createdTasks, err := sourcegraph.NewClientFromContext(ctx).Builds.CreateTasks(ctx, &sourcegraph.BuildsCreateTasksOp{
			Build: build.Spec(),
			Tasks: tasks,
		})
		if err != nil {
			return nil, err
		}
		states := make([]builder.TaskState, len(createdTasks.BuildTasks))
		for i, task := range createdTasks.BuildTasks {
			states[i] = &taskState{
				task: task.Spec(),
				log:  newLogger(task.Spec()),
			}
		}
		return states, nil
	}

	// FinalBuildConfig: Save config as BuilderConfig on the build.
	b.FinalBuildConfig = func(ctx context.Context, configYAML string) error {
		_, err := sourcegraph.NewClientFromContext(ctx).Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
			Build: build.Spec(),
			Info:  sourcegraph.BuildUpdate{BuilderConfig: string(configYAML)},
		})
		return err
	}

	return &b, nil
}

// getContainerAppURL gets the Sourcegraph server's app URL from the
// POV of the Docker containers.
func getContainerAppURL(ctx context.Context) (*url.URL, error) {
	cl := sourcegraph.NewClientFromContext(ctx)

	// Get the app URL from the POV of the Docker containers.
	serverConf, err := cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return nil, err
	}
	_, containerAppURLStr, err := containerAddrForHost(serverConf.AppURL)
	if err != nil {
		return nil, err
	}
	return url.Parse(containerAppURLStr)
}

// getSrclibImportURL constructs the srclib import URL to POST srclib
// data to, after the srclib build steps complete.
func getSrclibImportURL(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*url.URL, error) {
	srclibImportURL, err := httpapirouter.URL(httpapirouter.SrclibImport, repoRev.RouteVars())
	if err != nil {
		return nil, err
	}
	srclibImportURL.Path = "/.api" + srclibImportURL.Path

	containerAppURL, err := getContainerAppURL(ctx)
	if err != nil {
		return nil, err
	}

	return containerAppURL.ResolveReference(srclibImportURL), nil
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
func getCloneURLAndAuthForBuild(ctx context.Context, repo *sourcegraph.Repo) (cloneURL string, netrc *plugin.NetrcEntry, err error) {
	cloneURL = repo.HTTPCloneURL
	host, err := urlHostNoPort(cloneURL)
	if err != nil {
		return
	}

	switch {
	case repo.Mirror: // Mirror repos that live elsewhere.
		if repo.Private {
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

	case repo.Origin == "": // Repos hosted on this server.
		// NOTE: This assumes that if the if-condition below
		// holds, the repo's HTTPCloneURL is on the trusted server. If
		// it's ever possible for the HTTPCloneURL to be on a
		// different server but still have this if-condition hold,
		// then we could leak the user's credentials.
		netrc, err = getHostNetrcEntry(ctx)
		if err != nil {
			return
		}
		if len(netrc.Password) > 255 {
			// This should not occur anymore, but it is very
			// difficult to debug if it does, so log it
			// anyway.
			log15.Warn("warning: Long repository password is incompatible with git < 2.0. If you see git authentication errors, upgrade to git 2.0+.", "repo", repo.URI, "password length", len(netrc.Password))
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
func getHostNetrcEntry(ctx context.Context) (*plugin.NetrcEntry, error) {
	containerAppURL, err := getContainerAppURL(ctx)
	if err != nil {
		return nil, err
	}

	token := client.Credentials.GetAccessToken()
	if token == "" {
		return nil, errors.New("can't generate local netrc entry: token is empty")
	}
	host, _, err := containerAddrForHost(containerAppURL.String())
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
