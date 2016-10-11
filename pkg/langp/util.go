package langp

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/slimsag/untargz"

	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cmdutil"
)

var btrfsPresent bool

func init() {
	_, err := exec.LookPath("btrfs")
	if err == nil {
		btrfsPresent = true
	} else if os.Args[0] != "src" {
		log.Println("btrfs command not available, assuming filesystem is not btrfs")
	}
}

func btrfsSubvolumeCreate(ctx context.Context, path string) error {
	if !btrfsPresent {
		return os.Mkdir(path, 0700)
	}
	return CmdRun(ctx, exec.Command("btrfs", "subvolume", "create", path))
}

func btrfsSubvolumeSnapshot(ctx context.Context, subvolumePath, snapshotPath string) error {
	if !btrfsPresent {
		// TODO: This isn't portable outside *nix, but it does spare us a lot
		// of complex logic. Maybe find a good package to copy a directory.
		return CmdRun(ctx, exec.Command("cp", "-r", subvolumePath, snapshotPath))
	}
	return CmdRun(ctx, exec.Command("btrfs", "subvolume", "snapshot", subvolumePath, snapshotPath))
}

// dirExists tells if the directory p exists or not.
func dirExists(p string) (bool, error) {
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func lspKindToSymbol(kind lsp.SymbolKind) string {
	switch kind {
	case lsp.SKPackage:
		return "package"
	case lsp.SKField:
		return "field"
	case lsp.SKFunction:
		return "func"
	case lsp.SKMethod:
		return "method"
	case lsp.SKVariable:
		return "var"
	case lsp.SKClass:
		return "type"
	case lsp.SKInterface:
		return "interface"
	case lsp.SKConstant:
		return "const"
	default:
		// TODO(keegancsmith) We haven't implemented all types yet,
		// just what Go uses
		return "unknown"
	}
}

// ExpandSGPath expands the $SGPATH variable in the given string, except it
// uses ~/.sourcegraph as the default if $SGPATH is not set.
func ExpandSGPath(s string) (string, error) {
	sgpath := os.Getenv("SGPATH")
	if sgpath == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		sgpath = filepath.Join(u.HomeDir, ".sourcegraph")
	}
	return strings.Replace(s, "$SGPATH", sgpath, -1), nil
}

// RepoCloneURL returns a repo clone URL with authentication in it.
func RepoCloneURL(ctx context.Context, repo string) (cloneURI string) {
	token, _ := ctx.Value(GitHubTokenKey).(string)
	if token != "" && strings.HasPrefix(repo, "github.com/") {
		return fmt.Sprintf("https://x-oauth-token:%s@%s", token, repo)
	}
	return "https://" + repo
}

var repoAliases = []struct {
	// OldPrefix is the prefix of the import path to match, e.g. "golang.org/x/"
	OldPrefix string

	// NewPrefix is what to replace the OldPrefix with, e.g. "github.com/golang/"
	NewPrefix string
}{
	{
		OldPrefix: "github.com/slimsag/semver",
		NewPrefix: "azul3d.org/semver.v2",
	},
	{
		OldPrefix: "github.com/azul3d/",
		NewPrefix: "azul3d.org/",
	},
	{
		OldPrefix: "github.com/sourcegraph/sourcegraph",
		NewPrefix: "sourcegraph.com/sourcegraph/sourcegraph",
	},
	{
		// This special case needs to appear before the golang.org/x
		// since it is more specific. We further special case
		// github.com/golang/go so we want to preserve it.
		OldPrefix: "github.com/golang/go",
		NewPrefix: "github.com/golang/go",
	},
	{
		OldPrefix: "github.com/golang/",
		NewPrefix: "golang.org/x/",
	},
	{
		OldPrefix: "github.com/kubernetes/",
		NewPrefix: "k8s.io/",
	},
	{
		OldPrefix: "github.com/grpc/grpc-go",
		NewPrefix: "google.golang.org/grpc",
	},
	{
		OldPrefix: "github.com/GoogleCloudPlatform/google-cloud-go",
		NewPrefix: "cloud.google.com/go",
	},
	{
		OldPrefix: "github.com/google/google-api-go-client",
		NewPrefix: "google.golang.org/api",
	},
}

// ResolveRepoAlias returns import path of the given repository URI, it takes
// special care of sourcegraph main repository and others.
func ResolveRepoAlias(repo string) (importPath string) {
	for _, alias := range repoAliases {
		if strings.HasPrefix(repo, alias.OldPrefix) {
			return alias.NewPrefix + strings.TrimPrefix(repo, alias.OldPrefix)
		}
	}
	return repo
}

// UnresolveRepoAlias performs the opposite action of ResolveRepoAlias.
func UnresolveRepoAlias(repo string) string {
	for _, alias := range repoAliases {
		if strings.HasPrefix(repo, alias.NewPrefix) {
			return alias.OldPrefix + strings.TrimPrefix(repo, alias.NewPrefix)
		}
	}
	return repo
}

// CmdOutput is a helper around c.Output which logs the command, how long it
// took to run, and a nice error in the event of failure.
func CmdOutput(ctx context.Context, c *exec.Cmd) (stdout []byte, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, c.Args[0])
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()
	span.SetTag("command", strings.Join(c.Args, " "))
	span.SetTag("env", strings.Join(c.Env, "; "))

	start := time.Now()
	stdout, err = cmdutil.Output(c)
	log.Printf("TIME: %v '%s'\n", time.Since(start), strings.Join(c.Args, " "))
	if err != nil {
		log.Println(err)
		return stdout, err
	}
	return stdout, nil
}

// CmdRun is a helper around c.Run which logs the command, how long it took to
// run, and a nice error in the event of failure.
func CmdRun(ctx context.Context, c *exec.Cmd) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, c.Args[0])
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()
	span.SetTag("command", strings.Join(c.Args, " "))
	span.SetTag("env", strings.Join(c.Env, "; "))

	start := time.Now()
	err = cmdutil.Run(c)
	log.Printf("TIME: %v '%s'\n", time.Since(start), strings.Join(c.Args, " "))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// UpdateRepo updates the git repository in the specified directory to the
// specified revision.
func UpdateRepo(ctx context.Context, rev, dir string) error {
	// Update our repo to match the remote.
	cmd := exec.Command("git", "remote", "update", "--prune")
	cmd.Dir = dir
	if err := CmdRun(ctx, cmd); err != nil {
		return err
	}

	// Reset to the specific revision.
	cmd = exec.Command("git", "reset", "--hard", rev)
	cmd.Dir = dir
	return CmdRun(ctx, cmd)
}

var fastCloneClient = &http.Client{
	Timeout: 30 * time.Second,
}

// FastClone downloads a tarball archive of the specified repository at the
// given revision and extracts it to the destination directory. Once finished,
// the destination directory is not a proper git repository, but it can be
// later restored to one via RestoreRepo.
func FastClone(ctx context.Context, repoURI, rev, dir string) (err error) {
	start := time.Now()
	span, ctx := opentracing.StartSpanFromContext(ctx, "FastClone (GET + Extract tarball)")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
		}
		span.Finish()
	}()

	org, repo, err := splitGitHubCloneURI(repoURI)
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("TIME: FastClone %v %s/%s@%s\n", time.Since(start), org, repo, rev)
	}()

	tarball := &url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   fmt.Sprintf("repos/%s/%s/tarball/%s", org, repo, rev),
	}
	if token, _ := ctx.Value(GitHubTokenKey).(string); token != "" {
		tarball.RawQuery = "access_token=" + token
	}

	// Fetch and extract tarball to the destination.
	resp, err := fastCloneClient.Get(tarball.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	span.SetTag("status", resp.Status)
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		if len(body) > 1024 {
			body = body[:1024]
		}
		return fmt.Errorf("%s; body: '%s'", resp.Status, string(body))
	}
	err = untargz.Extract(resp.Body, dir, &untargz.Opts{
		TrimPathElements: 1, // Because GitHub archives always have a containing folder named "org-repo".
	})
	if err != nil {
		return err
	}

	// We fake having a full git repo by just putting the minimal amount
	// of information in to support:
	// * git rev-parse --show-toplevel HEAD
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err = CmdRun(ctx, cmd); err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, ".git/HEAD"), []byte(rev+"\n"), 0644)
	return err
}

// RestoreRepo restores the specified repoDir, which is assumed to be an
// extracted tarball of repository sources at the specified revision, back to
// it's full/complete natural state (i.e. as if you'd just cloned and
// `git reset --hard <rev>` the repository).
func RestoreRepo(ctx context.Context, cloneURI, rev, dir string) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RestoreRepo (tarball -> git repo)")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
		}
		span.Finish()
	}()

	cmd := exec.Command("git", "remote", "add", "origin", cloneURI)
	cmd.Dir = dir
	if err := CmdRun(ctx, cmd); err != nil {
		return err
	}
	cmd = exec.Command("git", "fetch")
	cmd.Dir = dir
	if err := CmdRun(ctx, cmd); err != nil {
		return err
	}
	cmd = exec.Command("git", "checkout", "-f", rev)
	cmd.Dir = dir
	return CmdRun(ctx, cmd)
}

func splitGitHubCloneURI(s string) (org, repo string, err error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", "", err
	}
	if u.Host != "github.com" {
		return "", "", fmt.Errorf("not a github.com clone URI (%q)", s)
	}
	parts := strings.Split(u.Path, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("FastClone: failed to parse repo URI %q, found parts %q", s, parts)
	}
	return parts[1], parts[2], nil
}
