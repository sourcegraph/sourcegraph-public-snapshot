package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/console/internal/webapp"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

const port = "3189"

//go:embed web/static
var staticFiles embed.FS
var staticFilesFS, _ = fs.Sub(staticFiles, "web/static")

func main() {
	env.Lock()
	env.HandleHelpFlag()

	liblog := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer liblog.Sync()

	trace.Init()
	profiler.Init()

	logger := log.Scoped("console", "")

	dsns, err := postgresdsn.DSNsBySchema([]string{"console"})
	if err != nil {
		logger.Fatal("failed to get PostgreSQL DSN", log.Error(err))
	}

	sqlDB, err := connections.EnsureNewConsoleDB(dsns["console"], "console", &observation.TestContext)
	if err != nil {
		logger.Fatal("failed to initialize database store", log.Error(err))
	}
	db := database.NewDB(logger, sqlDB)
	_ = db // TODO(sqs)

	host := ""
	if env.InsecureDev {
		host = "localhost"
	}

	// TODO(sqs)
	addr := net.JoinHostPort(host, port)
	externalURL, err := url.Parse("http://" + addr)
	if err != nil {
		logger.Fatal("unable to determine external URL", log.Error(err))
	}

	webapp := webapp.New(webapp.Config{
		ExternalURL: *externalURL,
		StaticFiles: staticFilesDevFS, // staticFilesFS, TODO(sqs): switch on some env var
		SessionKey:  "asdf",           // TODO(sqs) SECURITY(sqs)
	})
	webapp.Logger = logger

	logger.Info("listening", log.String("addr", addr))
	if err := http.ListenAndServe(addr, webapp); err != nil {
		logger.Fatal("failed to start HTTP listener", log.Error(err))
	}
}

var staticFilesDevFS esbuildFS

type esbuildFS struct{}

func (esbuildFS) Open(name string) (fs.File, error) {
	name = path.Clean(strings.TrimPrefix(name, "/"))
	if name == "." {
		name = ""
	}
	resp, err := http.Get("http://localhost:8077/" + name)
	fmt.Println("GET", "http://localhost:8077/"+name)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		var err error
		if resp.StatusCode == http.StatusNotFound {
			err = fs.ErrNotExist
		} else {
			err = fmt.Errorf("http status %d", resp.StatusCode)
		}
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return esbuildFSFile{name: name, Reader: bytes.NewReader(data)}, nil
}

type esbuildFSFile struct {
	name string
	*bytes.Reader
}

func (f esbuildFSFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (esbuildFSFile) Close() error { return nil }

func (fi esbuildFSFile) Name() string { return "/" + fi.name }
func (fi esbuildFSFile) Mode() fs.FileMode {
	if fi.name == "" {
		return fs.ModeDir
	}
	return 0
}
func (fi esbuildFSFile) ModTime() time.Time { return time.Time{} }
func (fi esbuildFSFile) IsDir() bool        { return fi.Mode()&fs.ModeDir != 0 }
func (fi esbuildFSFile) Sys() interface{}   { return nil }
