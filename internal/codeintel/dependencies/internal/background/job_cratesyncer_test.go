package background

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCrateSyncer(t *testing.T) {
	clock := glock.NewMockClock()
	rawConfig, _ := json.Marshal(schema.RustPackagesConnection{
		IndexRepositoryName:         "github.com/rust-lang/crates.io-index",
		IndexRepositorySyncInterval: "1m",
	})

	// dont need any functionality for this
	autoindexSvc := NewMockAutoIndexingService()

	refs := make(map[reposource.PackageName][]shared.MinimalPackageRepoRefVersion)
	dependenciesSvc := NewMockDependenciesService()
	dependenciesSvc.InsertPackageRepoRefsFunc.SetDefaultHook(func(ctx context.Context, refList []shared.MinimalPackageRepoRef) (newRef []shared.PackageRepoReference, newV []shared.PackageRepoRefVersion, err error) {
		for _, r := range refList {
			if versions, ok := refs[r.Name]; ok {
				refs[r.Name] = append(versions, r.Versions...)
				for _, v := range r.Versions {
					if slices.ContainsFunc(versions, func(v2 shared.MinimalPackageRepoRefVersion) bool {
						return v.Version == v2.Version && v.Blocked == v2.Blocked
					}) {
						newV = append(newV, shared.PackageRepoRefVersion{Version: v.Version})
					}
				}
			} else {
				newRef = append(newRef, shared.PackageRepoReference{Name: r.Name})
				for _, v := range r.Versions {
					newV = append(newV, shared.PackageRepoRefVersion{Version: v.Version})
				}
				refs[r.Name] = r.Versions
			}
		}
		return
	})

	gitclient := gitserver.NewMockClient()
	gitclient.LsFilesFunc.SetDefaultReturn([]string{"petgraph", "percent"}, nil)
	gitclient.ArchiveReaderFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
		var archive io.ReadCloser
		switch opts.Pathspecs[0] {
		case "petgraph":
			archive = createArchive(t, fileInfo{"petgraph", []byte(petgraphJSON)})
		case "percent":
			archive = createArchive(t, fileInfo{"percent", []byte(percentEncJSON)})
		}
		return archive, nil
	})

	extsvcStore := NewMockExternalServiceStore()
	extsvcStore.ListFunc.SetDefaultReturn([]*types.ExternalService{{
		ID:     1,
		Config: encryption.NewUnencrypted(string(rawConfig)),
	}}, nil)
	extsvcStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int64) (*types.ExternalService, error) {
		clock.Advance(time.Second)
		return &types.ExternalService{
			ID:         id,
			LastSyncAt: clock.Now(),
		}, nil
	})

	job := crateSyncerJob{
		archiveWindowSize: 1,
		autoindexingSvc:   autoindexSvc,
		dependenciesSvc:   dependenciesSvc,
		gitClient:         gitclient,
		extSvcStore:       extsvcStore,
		clock:             clock,
		operations:        newOperations(&observation.TestContext),
	}

	t.Run("Success", func(t *testing.T) {
		if err := job.handleCrateSyncer(context.Background(), time.Second); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dependenciesSvc.InsertPackageRepoRefsFunc.History()) != 2 {
			t.Errorf("unexpected number of calls to InsertPackageRepoRefs (want=%d, got=%d)", 2, len(dependenciesSvc.InsertPackageRepoRefsFunc.History()))
		}

		if len(extsvcStore.GetByIDFunc.History()) != 1 {
			t.Errorf("unexpected number of calls to GetByID (want=%d, got=%d)", 1, len(extsvcStore.GetByIDFunc.History()))
		}

		if len(autoindexSvc.QueueIndexesForPackageFunc.History()) != 6 {
			t.Errorf("unexpected number of calls to QueueIndexesForPackageFunc (want=%d, got=%d)", 6, len(autoindexSvc.QueueIndexesForPackageFunc.History()))
		}
	})

	t.Run("Fetch arhive err", func(t *testing.T) {
		dependenciesSvc.InsertPackageRepoRefsFunc.history = dependenciesSvc.InsertPackageRepoRefsFunc.history[:0]
		extsvcStore.GetByIDFunc.history = extsvcStore.GetByIDFunc.history[:0]
		autoindexSvc.QueueIndexesForPackageFunc.history = autoindexSvc.QueueIndexesForPackageFunc.history[:0]

		gitclient.ArchiveReaderFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
			if slices.Contains(opts.Pathspecs, "petgraph") {
				return createArchive(t, fileInfo{"petgraph", []byte(petgraphJSON)}), nil
			}
			return nil, errors.New("expected err")
		})

		const expectedErrString = `failed to git archive repo "github.com/rust-lang/crates.io-index": expected err`
		if err := job.handleCrateSyncer(context.Background(), time.Second); err == nil {
			t.Fatalf("unexpected nil error: %v", err)
		} else if err.Error() != expectedErrString {
			t.Fatalf("unexpected error (want=%q, got=%q)", expectedErrString, err)
		}

		if len(dependenciesSvc.InsertPackageRepoRefsFunc.History()) != 1 {
			t.Errorf("unexpected number of calls to InsertPackageRepoRefs (want=%d, got=%d)", 1, len(dependenciesSvc.InsertPackageRepoRefsFunc.History()))
		}

		if len(extsvcStore.GetByIDFunc.History()) != 0 {
			t.Errorf("unexpected number of calls to GetByID (want=%d, got=%d)", 0, len(extsvcStore.GetByIDFunc.History()))
		}

		if len(autoindexSvc.QueueIndexesForPackageFunc.History()) != 0 {
			t.Errorf("unexpected number of calls to QueueIndexesForPackageFunc (want=%d, got=%d)", 0, len(autoindexSvc.QueueIndexesForPackageFunc.History()))
		}
	})

	t.Run("Crate info JSON error", func(t *testing.T) {
		dependenciesSvc.InsertPackageRepoRefsFunc.history = dependenciesSvc.InsertPackageRepoRefsFunc.history[:0]
		extsvcStore.GetByIDFunc.history = extsvcStore.GetByIDFunc.history[:0]
		autoindexSvc.QueueIndexesForPackageFunc.history = autoindexSvc.QueueIndexesForPackageFunc.history[:0]

		gitclient.ArchiveReaderFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
			if slices.Contains(opts.Pathspecs, "petgraph") {
				return createArchive(t, fileInfo{"petgraph", []byte(petgraphJSON[:len(petgraphJSON)-5])}), nil
			}
			return createArchive(t, fileInfo{"percent", []byte(percentEncJSON)}), nil
		})

		const expectedErrString = `malformed crate info`
		if err := job.handleCrateSyncer(context.Background(), time.Second); err == nil {
			t.Fatalf("unexpected nil error: %v", err)
		} else if !strings.Contains(err.Error(), expectedErrString) {
			t.Fatalf("unexpected error (want contains=%q, got=%q)", expectedErrString, err)
		}

		if len(dependenciesSvc.InsertPackageRepoRefsFunc.History()) != 1 {
			t.Errorf("unexpected number of calls to InsertPackageRepoRefs (want=%d, got=%d)", 1, len(dependenciesSvc.InsertPackageRepoRefsFunc.History()))
		}

		if len(extsvcStore.GetByIDFunc.History()) != 1 {
			t.Errorf("unexpected number of calls to GetByID (want=%d, got=%d)", 1, len(extsvcStore.GetByIDFunc.History()))
		}

		if len(autoindexSvc.QueueIndexesForPackageFunc.History()) != 2 {
			t.Errorf("unexpected number of calls to QueueIndexesForPackageFunc (want=%d, got=%d)", 2, len(autoindexSvc.QueueIndexesForPackageFunc.History()))
		}
	})
}

func createArchive(t *testing.T, info fileInfo) io.ReadCloser {
	t.Helper()

	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	addFileToTarball(t, tarWriter, info)

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
