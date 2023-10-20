package background

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"
	"unsafe"

	"github.com/derision-test/glock"
	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type crateSyncerJob struct {
	archiveWindowSize int
	autoindexingSvc   AutoIndexingService
	dependenciesSvc   DependenciesService
	gitClient         gitserver.Client
	extSvcStore       ExternalServiceStore
	clock             glock.Clock
	operations        *operations
}

func NewCrateSyncer(
	observationCtx *observation.Context,
	autoindexingSvc AutoIndexingService,
	dependenciesSvc DependenciesService,
	gitClient gitserver.Client,
	extSvcStore ExternalServiceStore,
) goroutine.BackgroundRoutine {
	ctx := actor.WithInternalActor(context.Background())

	// By default, sync crates every 12h, but the user can customize this interval
	// through site-admin configuration of the RUSTPACKAGES code host.
	interval := time.Hour * 12
	_, externalService, _ := singleRustExternalService(ctx, extSvcStore)
	if externalService != nil {
		config, err := rustPackagesConfig(ctx, externalService)
		if err == nil { // silently ignore config errors.
			customInterval, err := time.ParseDuration(config.IndexRepositorySyncInterval)
			if err == nil { // silently ignore duration decoding error.
				interval = customInterval
			}
		}
	}

	job := crateSyncerJob{
		// average file size is ~10022bytes, 5000 files gives us an average (uncompressed) archive size of
		// about ~48MB. This will require ~21 gitserver archive calls
		archiveWindowSize: 5000,
		autoindexingSvc:   autoindexingSvc,
		dependenciesSvc:   dependenciesSvc,
		gitClient:         gitClient,
		extSvcStore:       extSvcStore,
		clock:             glock.NewRealClock(),
		operations:        newOperations(observationCtx),
	}

	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return job.handleCrateSyncer(ctx, interval)
		}),
		goroutine.WithName("codeintel.crates-syncer"),
		goroutine.WithDescription("syncs the crates list from the index to dependency repos table"),
		goroutine.WithInterval(interval),
	)
}

func (j *crateSyncerJob) handleCrateSyncer(ctx context.Context, interval time.Duration) (err error) {
	ctx, _, endObservation := j.operations.handleCrateSyncer.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	exists, externalService, err := singleRustExternalService(ctx, j.extSvcStore)
	if !exists || err != nil {
		// err can be nil when there is no RUSTPACKAGES code host.
		return err
	}

	config, err := rustPackagesConfig(ctx, externalService)
	if err != nil {
		return err
	}

	if config.IndexRepositoryName == "" {
		// Do nothing if the index repository is not configured.
		return nil
	}

	repoName := api.RepoName(config.IndexRepositoryName)

	// We should use an internal actor when doing cross service calls.
	clientCtx := actor.WithInternalActor(ctx)

	update, err := j.gitClient.RequestRepoUpdate(clientCtx, repoName, interval)
	if err != nil {
		return err
	}
	if update != nil && update.Error != "" {
		return errors.Newf("failed to update repo %s, error %s", repoName, update.Error)
	}

	allFilesStr, err := j.gitClient.LsFiles(ctx, repoName, "HEAD")
	if err != nil {
		return err
	}
	// safe according to rule #1 of pkg/unsafe
	allFiles := *(*[]gitdomain.Pathspec)(unsafe.Pointer(&allFilesStr))

	var (
		allCratePkgs       []shared.MinimalPackageRepoRef
		didInsertNewCrates bool
		// we dont want to throw away all work if we only read
		// the crates index partially
		cratesReadErr error
	)

	for len(allFiles) > 0 {
		var batch []gitdomain.Pathspec
		if len(allFiles) <= j.archiveWindowSize {
			batch, allFiles = allFiles, nil
		} else {
			batch, allFiles = allFiles[:j.archiveWindowSize], allFiles[j.archiveWindowSize:]
		}

		buf, err := j.readIndexArchiveBatch(clientCtx, repoName, batch)
		if err != nil {
			return err
		}

		tr := tar.NewReader(buf)
		if err != nil {
			return err
		}

		for {
			header, err := tr.Next()
			if err != nil {
				if err != io.EOF {
					cratesReadErr = errors.Append(cratesReadErr, err)
					break
				}
				break
			}

			// Skip directory entries
			if strings.HasSuffix(header.Name, "/") {
				continue
			}

			// `.github/` contains non-crates information
			if strings.HasPrefix(header.Name, ".github") {
				continue
			}

			// `config.json` contains metadata about the repo,
			// we can use this file later if we want to support other
			// file formats
			if header.Name == "config.json" {
				continue
			}

			buf := bytes.NewBuffer(make([]byte, 0, header.Size))
			if _, err := io.CopyN(buf, tr, header.Size); err != nil {
				cratesReadErr = errors.Append(cratesReadErr, err)
				break
			}

			pkgs, err := parseCrateInformation(buf.Bytes())
			if err != nil {
				cratesReadErr = errors.Append(cratesReadErr, err)
				break
			}

			allCratePkgs = append(allCratePkgs, pkgs...)

			newCrates, newVersions, err := j.dependenciesSvc.InsertPackageRepoRefs(ctx, pkgs)
			if err != nil {
				return errors.Wrapf(err, "failed to insert rust crate")
			}
			didInsertNewCrates = didInsertNewCrates || len(newCrates) != 0 || len(newVersions) != 0
		}
	}

	nextSync := j.clock.Now()
	if didInsertNewCrates {
		// We picked up new crates so we trigger a new sync for the RUSTPACKAGES code host.
		externalService.NextSyncAt = nextSync
		if err := j.extSvcStore.Upsert(ctx, externalService); err != nil {
			return errors.Append(cratesReadErr, err)
		}

		for attemptsRemaining := 5; attemptsRemaining > 0; attemptsRemaining-- {
			externalService, err = j.extSvcStore.GetByID(ctx, externalService.ID)
			if err != nil && attemptsRemaining == 0 {
				return errors.Append(cratesReadErr, err)
			} else if err != nil || !externalService.LastSyncAt.After(nextSync) {
				// mirrors backoff in job_dependency_indexing_scheduler.go
				j.clock.Sleep(time.Second * 30)
				continue
			}

			break
		}

		var queueErrs errors.MultiError
		for _, pkg := range allCratePkgs {
			for _, version := range pkg.Versions {
				if err := j.autoindexingSvc.QueueIndexesForPackage(clientCtx, shared.MinimialVersionedPackageRepo{
					Scheme:  pkg.Scheme,
					Name:    pkg.Name,
					Version: version.Version,
				}); err != nil {
					queueErrs = errors.Append(queueErrs, err)
				}
			}
		}

		return errors.Append(cratesReadErr, queueErrs)
	}

	return cratesReadErr
}

func (j *crateSyncerJob) readIndexArchiveBatch(ctx context.Context, repoName api.RepoName, batch []gitdomain.Pathspec) (io.Reader, error) {
	reader, err := j.gitClient.ArchiveReader(
		ctx,
		repoName,
		gitserver.ArchiveOptions{
			Treeish:   "HEAD",
			Format:    gitserver.ArchiveFormatTar,
			Pathspecs: batch,
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to git archive repo %q", repoName)
	}
	// important to read this into memory asap
	defer reader.Close()

	// read into mem to avoid holding connection open, with a 50MB buffer
	buf := bytes.NewBuffer(make([]byte, 0, 50*1024*1024))
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, errors.Wrap(err, "failed to read git archive")
	}

	return bytes.NewReader(buf.Bytes()), nil
}

// rustPackagesConfig returns the configuration for the provided RUSTPACKAGES code host.
func rustPackagesConfig(ctx context.Context, externalService *dbtypes.ExternalService) (*schema.RustPackagesConnection, error) {
	rawConfig, err := externalService.Config.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	config := &schema.RustPackagesConnection{}
	normalized, err := jsonc.Parse(rawConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse JSON config for rust external service %s", rawConfig)
	}

	if err = jsoniter.Unmarshal(normalized, config); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal rust external service config %s", rawConfig)
	}
	return config, nil
}

// singleRustExternalService returns the single external service type with kind RUSTPACKAGES.
// The external service and the error are both nil when there are no RUSTPACKAGES code hosts.
// The `exists` return value is false whenever externalService is nil, and it exists only as a
// reminder that `nil, nil` is a valid return value (no external service, no error).
func singleRustExternalService(ctx context.Context, store ExternalServiceStore) (exists bool, externalService *dbtypes.ExternalService, err error) {
	kind := extsvc.KindRustPackages

	externalServices, err := store.List(ctx, database.ExternalServicesListOptions{
		Kinds: []string{kind},
	})
	if err != nil {
		return false, nil, errors.Wrapf(err, "failed to list rust external service types")
	}

	//  Skip if RUSTPACKAGES not enabled
	if len(externalServices) == 0 {
		return false, nil, nil
	}

	//  We only support having a single RUSTPACKAGES external service type, for now
	if len(externalServices) > 1 {
		return false, nil, errors.Errorf("multiple external services with kind %s", kind)
	}

	return true, externalServices[0], nil
}

// parseCrateInformation parses the newline-delimited JSON file for a crate,
// assuming the pattern that's used in the github.com/rust-lang/crates.io-index
func parseCrateInformation(contents []byte) ([]shared.MinimalPackageRepoRef, error) {
	result := make([]shared.MinimalPackageRepoRef, 0, 1)

	instant := time.Now()

	lr := byteutils.NewLineReader(contents)

	for lr.Scan() {
		line := lr.Line()

		if len(line) == 0 {
			continue
		}

		type crateInfo struct {
			Name    string `json:"name"`
			Version string `json:"vers"`
		}
		var info crateInfo
		err := json.Unmarshal(line, &info)
		if err != nil {
			return nil, errors.Wrapf(err, "malformed crate info (%q)", line)
		}

		name := reposource.PackageName(info.Name)
		result = append(result, shared.MinimalPackageRepoRef{
			Scheme: shared.RustPackagesScheme,
			Name:   name,
			// doing a bit of a dot-com specific assumption here, that all these packages are resolvable
			// and not covered by a filter.
			Versions:      []shared.MinimalPackageRepoRefVersion{{Version: info.Version, LastCheckedAt: &instant}},
			LastCheckedAt: &instant,
		})
	}

	return result, nil
}
