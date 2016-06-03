package worker

import (
	"errors"
	"net"
	"net/url"
	"path"
	"strings"

	droneexec "github.com/drone/drone-exec/exec"
	"github.com/drone/drone-plugin-go/plugin"
	"github.com/whilp/git-urls"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dockerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	httpapirouter "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker/builder"
	"sourcegraph.com/sqs/pbtypes"
)

func configureBuild(ctx context.Context, build *sourcegraph.BuildJob) (*builder.Builder, error) {
	var b builder.Builder

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	repoRev := sourcegraph.RepoRevSpec{
		Repo:     build.Spec.Repo.URI,
		CommitID: build.CommitID,
	}
	repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: repoRev.Repo})
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

	appURL, err := getAppURL(ctx)
	if err != nil {
		return nil, err
	}

	// Adjust the URL in case it uses localhost to an absolute IP.
	hostname, containerAppURL, err := containerAddrForHost(*appURL)
	if err != nil {
		return nil, err
	}

	containerCloneURL := *containerAppURL
	containerCloneURL.Path = repo.URI

	repoURL, err := parseCloneURL(repo)
	if err != nil {
		return nil, err
	}

	repoLink, err := droneRepoLink(*repoURL)
	if err != nil {
		return nil, err
	}

	b.Payload.Repo = &plugin.Repo{
		FullName:  build.Spec.Repo.URI,
		Clone:     containerCloneURL.String(),
		Link:      repoLink,
		IsPrivate: true,
		IsTrusted: true,
	}

	// Get the netrc entry we need for srclib-importing (and otherwise
	// contacting the Sourcegraph server from the container). These
	// may be the same credentials as the clone netrc credentials, but
	// that's not true in all cases (e.g., clone credentials could be
	// for GitHub).
	b.Payload.Netrc = append(b.Payload.Netrc, &plugin.NetrcEntry{
		Machine:  hostname,
		Login:    "x-oauth-basic",
		Password: build.AccessToken,
	})

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
		Notify: false,
		Debug:  true,
	}

	// SrclibImportURL
	srclibImportURL, err := getSrclibImportURL(ctx, repoRev, *containerAppURL)
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
		c, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			return nil, err
		}
		createdTasks, err := c.Builds.CreateTasks(ctx, &sourcegraph.BuildsCreateTasksOp{
			Build: build.Spec,
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
		c, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			return err
		}
		_, err = c.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
			Build: build.Spec,
			Info:  sourcegraph.BuildUpdate{BuilderConfig: string(configYAML)},
		})
		return err
	}

	return &b, nil
}

// getContainerAppURL gets the Sourcegraph server's app URL from the
// POV of the Docker containers.
func getAppURL(ctx context.Context) (*url.URL, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the app URL from the POV of the Docker containers.
	serverConf, err := cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return nil, err
	}

	return url.Parse(serverConf.AppURL)
}

// getSrclibImportURL constructs the srclib import URL to POST srclib
// data to, after the srclib build steps complete.
func getSrclibImportURL(ctx context.Context, repoRev sourcegraph.RepoRevSpec, containerAppURL url.URL) (*url.URL, error) {
	srclibImportURL, err := httpapirouter.URL(httpapirouter.SrclibImport, routevar.RepoRevRouteVars(routevar.RepoRev{Repo: repoRev.Repo, Rev: repoRev.CommitID}))
	if err != nil {
		return nil, err
	}
	srclibImportURL.Path = "/.api" + srclibImportURL.Path

	return containerAppURL.ResolveReference(srclibImportURL), nil
}

// parseCloneURL parses any valid HTTP or SSH git repo URL and normalizes it
// to a normal URL. For example:
//
// git@github.com:sourcegraph/srclib.git --> ssh://git@github.com/sourcegraph/srclib.git
func parseCloneURL(repo *sourcegraph.Repo) (*url.URL, error) {
	if repo.HTTPCloneURL != "" {
		return giturls.Parse(repo.HTTPCloneURL)
	} else if repo.SSHCloneURL != "" {
		return giturls.Parse(repo.SSHCloneURL)
	} else {
		return nil, errors.New("Must provide either an HTTP(S) or SSH clone URL")
	}
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
func droneRepoLink(u url.URL) (string, error) {
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
func containerAddrForHost(u url.URL) (string, *url.URL, error) {
	hostname := u.Host
	if strings.Contains(u.Host, ":") {
		var err error
		hostname, _, err = net.SplitHostPort(hostname)
		if err != nil {
			return "", nil, err
		}
	}

	if hostname == "localhost" {
		containerHostname, err := dockerutil.ContainerHost()
		if err != nil {
			return "", nil, err
		}
		hostname = containerHostname
		u.Host = strings.Replace(u.Host, "localhost", containerHostname, 1)
	}

	return hostname, &u, nil
}
