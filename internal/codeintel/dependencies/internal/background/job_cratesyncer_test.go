package background

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/fs"
	"path"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCrateSyncer(t *testing.T) {
	autoindexSvc := NewMockAutoIndexingService()
	dependenciesSvc := NewMockDependenciesService()
	gitclient := NewMockGitserverClient()
	gitclient.LsFilesFunc.SetDefaultReturn([]string{"petgraph", "percent"}, nil)
	gitclient.ArchiveReaderFunc.SetDefaultHook(func(ctx context.Context, sub authz.SubRepoPermissionChecker, name api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
		return createArchive(t, string(opts.Pathspecs[0])), nil
	})
	extsvcStore := NewMockExternalServiceStore()

	job := crateSyncerJob{
		archiveWindowSize: 1,
		autoindexingSvc:   autoindexSvc,
		dependenciesSvc:   dependenciesSvc,
		gitClient:         gitclient,
		extSvcStore:       extsvcStore,
		operations:        newOperations(&observation.TestContext),
	}

	job.handleCrateSyncer(context.Background(), time.Second)
}

func createArchive(t *testing.T, path string) io.ReadCloser {
	t.Helper()

	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	switch path {
	case "petgraph":
		addFileToTarball(t, tarWriter, fileInfo{"petgraph", []byte(petgraphJSON)})
	case "percent":
		addFileToTarball(t, tarWriter, fileInfo{"percent", []byte(percentEncJSON)})
	}

	return io.NopCloser(&buf)
}

func addFileToTarball(t *testing.T, tarWriter *tar.Writer, info fileInfo) error {
	t.Helper()
	header, err := tar.FileInfoHeader(&info, "")
	if err != nil {
		return err
	}
	header.Name = info.path
	if err = tarWriter.WriteHeader(header); err != nil {
		return errors.Wrapf(err, "failed to write header for %s", info.path)
	}
	_, err = tarWriter.Write(info.contents)
	return err
}

type fileInfo struct {
	path     string
	contents []byte
}

var _ fs.FileInfo = &fileInfo{}

func (info *fileInfo) Name() string       { return path.Base(info.path) }
func (info *fileInfo) Size() int64        { return int64(len(info.contents)) }
func (info *fileInfo) Mode() fs.FileMode  { return 0o600 }
func (info *fileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (info *fileInfo) IsDir() bool        { return false }
func (info *fileInfo) Sys() any           { return nil }

const petgraphJSON = `{"name":"petgraph","vers":"0.0.1","deps":[],"cksum":"cdf41894260194c9c6ef2286db9889f1d32510fb891570001e99e4b56945ad92","features":{},"yanked":false}
{"name":"petgraph","vers":"0.0.7","deps":[{"name":"rand","req":"*","features":[],"optional":false,"default_features":true,"target":null,"kind":"dev"}],"cksum":"12ae2a8781008c2d66bd98cb45db23b1631c6a2dc9d50c445d5f31e700cd8f66","features":{},"yanked":false}
{"name":"petgraph","vers":"0.1.0","deps":[{"name":"fixedbitset","req":"*","features":[],"optional":false,"default_features":true,"target":null,"kind":"normal"},{"name":"rand","req":"*","features":[],"optional":false,"default_features":true,"target":null,"kind":"dev"}],"cksum":"2c82ef6f7153886108ebfa52c83a536eb1ab575c274e57d098fa366510c7ed44","features":{},"yanked":false}
{"name":"petgraph","vers":"0.3.0-alpha.0","deps":[{"name":"fixedbitset","req":"^0.1.0","features":[],"optional":false,"default_features":true,"target":null,"kind":"normal"},{"name":"itertools","req":"^0.5","features":[],"optional":false,"default_features":true,"target":null,"kind":"normal"},{"name":"quickcheck","req":"^0.3","features":[],"optional":true,"default_features":true,"target":null,"kind":"normal"},{"name":"rand","req":"^0.3","features":[],"optional":false,"default_features":true,"target":null,"kind":"dev"}],"cksum":"142ae98dbb3bd0d90f86dafa9ac38824404923d9ff65e9530291a07557044237","features":{"all":["test","unstable","quickcheck"],"default":["stable_graph"],"generate":[],"stable_graph":[],"test":[],"unstable":["generate"]},"yanked":false}
`

const percentEncJSON = `{"name":"percent-encoding","vers":"1.0.0","deps":[{"name":"rustc-serialize","req":"^0.3","features":[],"optional":false,"default_features":true,"target":null,"kind":"dev"},{"name":"rustc-test","req":"^0.1","features":[],"optional":false,"default_features":true,"target":null,"kind":"dev"}],"cksum":"de154f638187706bde41d9b4738748933d64e6b37bdbffc0b47a97d16a6ae356","features":{},"yanked":false}
{"name":"percent-encoding","vers":"2.2.0","deps":[],"cksum":"478c572c3d73181ff3c2539045f6eb99e5491218eae919370993b890cdbdd98e","features":{"alloc":[],"default":["alloc"]},"yanked":false}
`
