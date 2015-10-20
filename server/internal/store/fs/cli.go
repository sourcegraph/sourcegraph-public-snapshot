package fs

import (
	"log"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
	sgxcli "src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/store/cli"

	_ "src.sourcegraph.com/sourcegraph/sgx"
)

var Stores = store.Stores{
	Accounts:          &Accounts{},
	Authorizations:    &Authorizations{},
	BuildLogs:         &BuildLogs{},
	Builds:            NewBuildStore(),
	Discussions:       &Discussions{},
	Password:          &Password{},
	RegisteredClients: &RegisteredClients{},
	RepoStatuses:      &RepoStatuses{},
	RepoVCS:           &RepoVCS{},
	Repos:             &Repos{},
	Users:             &Users{},
	UserKeys:          NewUserKeys(),
	Changesets:        &Changesets{},
}

func init() {
	cli.RegisterStores("fs", &Stores)
}

// newVFS creates a read-write VFS rooted at base, which can be either
// a local directory or a HTTP URL (pointing to an HTTP VFS
// server). It calls log.Fatal if it encounters an error because it's
// intended to be used at init time.
func newVFS(base string) rwvfs.FileSystem {
	if base == "" {
		// Will result in a panic when attempted to be used, which is
		// OK since this should not be empty.
		return nil
	}

	var fs rwvfs.FileSystem

	u, err := url.Parse(base)
	if err == nil && u.Scheme != "" {
		// base is a URL pointing to an HTTP VFS server.
		fs = rwvfs.HTTP(u, nil)
	} else {
		// base is a local directory.
		fs = rwvfs.OS(filepath.Clean(base))
	}

	setCreateParentDirs(fs)

	if err := rwvfs.MkdirAll(fs, "."); err != nil {
		log.Fatal(err)
	}

	return fs
}

func setCreateParentDirs(fs rwvfs.FileSystem) {
	if fs, ok := fs.(interface {
		CreateParentDirs(bool)
	}); ok {
		fs.CreateParentDirs(true)
	}
}

func init() {
	sgxcli.PostInit = append(sgxcli.PostInit, func() {
		_, err := sgxcli.Serve.AddGroup("Local filesystem storage (fs store)", "Local filesystem storage", &ActiveFlags)
		if err != nil {
			log.Fatal(err)
		}
	})

	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		ActiveFlags.Expand()

		// Construct filesystems once at init time so we can reuse
		// HTTP clients backing HTTP VFSes and avoid needless garbage.
		dbVFS := newVFS(ActiveFlags.DBDir)
		buildStoreVFS := newVFS(ActiveFlags.BuildStoreDir)
		repoStatusVFS := newVFS(ActiveFlags.RepoStatusDir)

		serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
			if dir := ActiveFlags.ReposDir; dir != "" {
				ctx = WithReposVFS(ctx, filepath.Clean(dir))
			}
			ctx = WithBuildStoreVFS(ctx, buildStoreVFS)
			ctx = WithDBVFS(ctx, dbVFS)
			ctx = WithRepoStatusVFS(ctx, repoStatusVFS)
			return ctx, nil
		})
	})
}

type Flags struct {
	ReposDir      string `long:"fs.repos-dir" description:"root dir containing repos" default:"$SGPATH/repos"`
	BuildStoreDir string `long:"fs.build-store-dir" description:"root dir (or HTTP VFS base URL) containing builds" default:"$SGPATH/buildstore"`
	DBDir         string `long:"fs.db-dir" description:"root dir containing user/account/etc. data" default:"$SGPATH/db"`
	RepoStatusDir string `long:"fs.repo-status-dir" description:"root dir containing repo statuses" default:"$SGPATH/statuses"`
}

var ActiveFlags Flags

func (f *Flags) Expand() {
	f.ReposDir = os.ExpandEnv(f.ReposDir)
	f.BuildStoreDir = os.ExpandEnv(f.BuildStoreDir)
	f.DBDir = os.ExpandEnv(f.DBDir)
	f.RepoStatusDir = os.ExpandEnv(f.RepoStatusDir)
}
