package sgx

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/kr/fs"
	"github.com/shurcooL/go/vfs/godocfs/godocfs"
	"golang.org/x/tools/godoc/vfs"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/misc/sampledata"
)

// prepareInitialOnboarding adds sample repos to show off
// Sourcegraph's capabilities on a newly set up server.
func (c *ServeCmd) prepareInitialOnboarding(ctx context.Context) error {
	if c.NoInitialOnboarding {
		return nil
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	type op struct {
		sourcegraph.ReposCreateOp
		pushRefspecs []string
		after        func(*sourcegraph.Repo, context.Context) error
	}
	ops := []op{
		{
			ReposCreateOp: sourcegraph.ReposCreateOp{
				URI:         "sample/golang/hello",
				VCS:         "git",
				Description: "A Go starter project to demonstrate Sourcegraph",
				Language:    "Go",
			},
			pushRefspecs: []string{"master"},
		},
	}

	do := func(ctx context.Context, op op) error {
		log15.Debug(fmt.Sprintf("Creating sample repo %s...", op.URI))
		repo, err := cl.Repos.Create(ctx, &op.ReposCreateOp)
		if err != nil {
			if grpc.Code(err) == codes.Unimplemented {
				log15.Info("Skipping creation of sample demonstration repo", "repo", op.URI)
				return nil
			}
			return err
		}

		dir, err := ioutil.TempDir("", "sourcegraph-sample-"+strings.Replace(repo.URI, "/", "-", -1))
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)

		// Write the sample repository data to disk (from the
		// sampledata VFS baked into this binary).
		vfsDir := path.Join("/repos", strings.TrimPrefix(repo.URI, "sample/"))
		sampledataVFS := walkableFileSystem{godocfs.New(sampledata.Data)}
		w := fs.WalkFS(vfsDir, sampledataVFS)
		for w.Step() {
			if err := w.Err(); err != nil {
				return err
			}

			if !w.Stat().Mode().IsRegular() {
				continue
			}

			path := strings.TrimPrefix(w.Path(), vfsDir+"/")
			path = strings.Replace(path, ".git-versioned", ".git", 1)
			if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(path)), 0700); err != nil {
				return err
			}

			fileData, err := vfs.ReadFile(sampledataVFS, w.Path())
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(filepath.Join(dir, path), fileData, 0600); err != nil {
				return err
			}
		}

		// Git push to the repo.
		{
			authedURL, err := authutil.AddAuthToURL(ctx, repo.HTTPCloneURL)
			if err != nil {
				return err
			}
			cmd := exec.Command("git", "push", authedURL)
			cmd.Args = append(cmd.Args, op.pushRefspecs...)
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				log15.Error("Git push of sample repo failed", "command", cmd.Args, "err", err, "output", string(out))
				return fmt.Errorf("command %v failed (%s)", cmd.Args, err)
			}
		}

		if op.after != nil {
			if err := op.after(repo, ctx); err != nil {
				return err
			}
		}

		log15.Debug(fmt.Sprintf("Created sample repo %s", repo.URI))
		return nil
	}
	for _, op := range ops {
		if err := do(ctx, op); err != nil {
			return err
		}
	}

	return nil
}
