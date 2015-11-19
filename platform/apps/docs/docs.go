// Package docs is a repository app that renders a static
// [Hugo](http://gohugo.io/) site defined within the repository.
//
// NOTE: It relies on a fork of Hugo, github.com/sqs/hugo@vfs, that
// removes dependencies on the OS filesystem and routes all FS access
// through a VFS.
package docs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
	"github.com/spf13/hugo/commands"
	"github.com/spf13/hugo/hugofs"
	"github.com/spf13/hugo/hugolib"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	platform.RegisterFrame(platform.RepoFrame{
		ID:      "docs",
		Title:   "Docs",
		Icon:    "book",
		Handler: http.HandlerFunc(handler),
	})

	commands.LoadDefaultSettings()
}

var (
	hugoGlobalMu sync.Mutex // guards hugo globals (e.g., hugofs.DestinationFS)

	cacheMu sync.Mutex                               // guards the following field
	cache   = map[sourcegraph.RepoRevSpec]afero.Fs{} // caches built static sites
)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.FromRequest(r)
	cl := sourcegraph.NewClientFromContext(ctx)

	repoRev, exists := pctx.RepoRevSpec(ctx)
	if !exists {
		http.Error(w, "could not parse repository spec from URL", http.StatusBadRequest)
		return
	}
	if len(repoRev.CommitID) != 40 {
		commit, err := cl.Repos.GetCommit(ctx, &repoRev)
		if err != nil {
			http.Error(w, "GetCommit: "+err.Error(), http.StatusInternalServerError)
			return
		}
		repoRev.CommitID = string(commit.ID)
	}

	// Hugo uses globals, like hugofs.DestinationFS, so we can only
	// run one of these handlers at a time.
	hugoGlobalMu.Lock()
	defer hugoGlobalMu.Unlock()

	fs, err := cachedBuild(ctx, repoRev)
	if err != nil {
		http.Error(w, err.Error(), errcode.HTTP(err))
		return
	}

	rw := httptest.NewRecorder()
	rw.Body = new(bytes.Buffer)

	fileserver := http.FileServer((&afero.HttpFs{SourceFs: fs}).Dir("."))
	fileserver.ServeHTTP(rw, r)
	rw.Header().Del("content-length")
	for k, vs := range rw.HeaderMap {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if _, err := rw.Body.WriteTo(w); err != nil {
		http.Error(w, "WriteTo: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// localHugoDir is a development mode switch that makes Hugo build docs
// from the specified dir on the local filesystem, not from the git
// repository. It is useful when you want to see live previews while
// editing docs.
var localHugoDir = os.Getenv("DEV_LOCAL_HUGO_DIR")

func getSourceFS(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (hugoDir string, fs afero.Fs, err error) {
	if localHugoDir != "" {
		return localHugoDir, afero.OsFs{}, nil
	}

	// TODO(sqs): Assumes that repo is on local disk. To remove this
	// assumption, we would need a VFS interface that operates over
	// gRPC to the RepoTree service.
	vcsRepo, err := vcs.Open("git", filepath.Join(os.Getenv("SGPATH"), "repos", repoRev.URI))
	if err != nil {
		return "", nil, err
	}
	vfs, err := vcsRepo.FileSystem(vcs.CommitID(repoRev.CommitID))
	if err != nil {
		return "", nil, err
	}
	fs = aferoVFS{vfs}

	hugoDir, err = findHugoDataDir(ctx, repoRev)
	if err != nil {
		var msg string
		if grpc.Code(err) == codes.NotFound {
			msg = fmt.Sprintf("No Hugo config.toml file found in %s", hugoDir)
		} else {
			msg = "Error configuring Hugo static site generator"
		}
		return "", nil, errors.New(msg)
	}

	return hugoDir, fs, nil
}

func cachedBuild(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (afero.Fs, error) {
	// Bypass cache if we are running in local dev mode.
	useCache := localHugoDir == ""

	cacheMu.Lock()
	defer cacheMu.Unlock()

	if useCache {
		fs, present := cache[repoRev]
		if present {
			return fs, nil
		}
	}

	fs, err := build(ctx, repoRev)
	if err != nil {
		return nil, err
	}

	if useCache {
		cache[repoRev] = fs
	}

	return fs, nil
}

func build(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (afero.Fs, error) {
	hugoDir, fs, err := getSourceFS(ctx, repoRev)
	if err != nil {
		return nil, err
	}

	hugofs.SourceFs = fs
	hugofs.DestinationFS = &afero.MemMapFs{}
	hugofs.OsFs = nil

	commands.Source = hugoDir
	commands.CfgFile = filepath.Join(hugoDir, "config.toml")

	viper.SetDefault("DataDir", hugoDir)
	viper.SetDefault("LayoutDir", filepath.Join(hugoDir, "layouts"))
	viper.SetDefault("ArchetypeDir", filepath.Join(hugoDir, "archetype"))
	viper.SetDefault("PublishDir", "")
	viper.SetDefault("StaticDir", filepath.Join(hugoDir, "static"))
	viper.SetDefault("ContentDir", filepath.Join(hugoDir, "content"))
	viper.SetDefault("Verbose", true)
	viper.Set("BuildDrafts", true)
	viper.SetDefault("BaseURL", pctx.BaseURI(ctx))

	configFile, err := hugofs.SourceFs.Open(commands.CfgFile)
	if err != nil {
		return nil, err
	}
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(configFile); err != nil {
		return nil, err
	}

	site := &hugolib.Site{}
	if err := site.Build(); err != nil {
		return nil, err
	}

	return hugofs.DestinationFS, nil
}

// findHugoDataDir determines the Hugo data directory (the one
// containing the config.toml file).
func findHugoDataDir(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (dir string, err error) {
	cl := sourcegraph.NewClientFromContext(ctx)

	// HACK: Look in a hacky config file called ".sourcegraph" for the
	// dir.
	configFile, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{RepoRev: repoRev, Path: ".sourcegraph"},
	})
	if err == nil {
		var config struct{ HugoDir string }
		if err := json.Unmarshal(configFile.Contents, &config); err != nil {
			return "", err
		}
		dir = filepath.Clean(config.HugoDir)
	} else if grpc.Code(err) == codes.NotFound {
		dir = "." // default
	} else if err != nil {
		return "", err
	}

	// Check that the dir actually exists and has a config.toml file.
	_, err = cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{RepoRev: repoRev, Path: filepath.Join(dir, "config.toml")},
	})
	return dir, err
}
