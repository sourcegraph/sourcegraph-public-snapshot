package fs

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
	sgxcli "src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	sgxcli.PostInit = append(sgxcli.PostInit, func() {
		_, err := sgxcli.Serve.AddGroup("Local filesystem storage", "Local filesystem storage", &activeFlags)
		if err != nil {
			log.Fatal(err)
		}
	})

	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		activeFlags.Expand()
		serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
			if dir := activeFlags.ReposDir; dir != "" {
				ctx = WithReposVFS(ctx, filepath.Clean(dir))
			}
			return ctx, nil
		})
	})
}

type Flags struct {
	ReposDir      string `long:"fs.repos-dir" description:"root dir containing repos" default:"$SGPATH/repos"`
	GitRepoMirror string `long:"fs.git-repo-mirror" description:"comma-separated string map in the form '<LocalRepoURI>:<GitRemoteURL>' defining which repos to mirror on a remote host"`
}

var activeFlags Flags

func (f *Flags) Expand() {
	f.ReposDir = os.ExpandEnv(f.ReposDir)
}
