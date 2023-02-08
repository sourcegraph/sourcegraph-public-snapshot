package background

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
	dependenciesSvc DependenciesService
	gitClient       GitserverClient
	extSvcStore     ExternalServiceStore
	operations      *operations
}

func NewCrateSyncer(
	observationCtx *observation.Context,
	dependenciesSvc DependenciesService,
	gitClient GitserverClient,
	extSvcStore ExternalServiceStore,
) goroutine.BackgroundRoutine {
	// By default, sync crates every 12h, but the user can customize this interval
	// through site-admin configuration of the RUSTPACKAGES code host.
	interval := time.Hour * 12
	_, externalService, _ := singleRustExternalService(context.Background(), extSvcStore)
	if externalService != nil {
		config, err := rustPackagesConfig(context.Background(), externalService)
		if err == nil { // silently ignore config errors.
			customInterval, err := time.ParseDuration(config.IndexRepositorySyncInterval)
			if err == nil { // silently ignore duration decoding error.
				interval = customInterval
			}
		}
	}

	job := crateSyncerJob{
		dependenciesSvc: dependenciesSvc,
		gitClient:       gitClient,
		extSvcStore:     extSvcStore,
		operations:      newOperations(observationCtx),
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.crates-syncer", "syncs the crates list from the index to dependency repos table",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return job.handleCrateSyncer(ctx, interval)
		}),
	)
}

func (b *crateSyncerJob) handleCrateSyncer(ctx context.Context, interval time.Duration) (err error) {
	ctx, _, endObservation := b.operations.handleCrateSyncer.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	exists, externalService, err := singleRustExternalService(ctx, b.extSvcStore)
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

	update, err := b.gitClient.RequestRepoUpdate(clientCtx, repoName, interval)
	if err != nil {
		return err
	}
	if update != nil && update.Error != "" {
		return errors.Newf("failed to update repo %s, error %s", repoName, update.Error)
	}
	reader, err := b.gitClient.ArchiveReader(
		clientCtx,
		nil,
		repoName,
		gitserver.ArchiveOptions{
			Treeish:   "HEAD",
			Format:    gitserver.ArchiveFormatTar,
			Pathspecs: []gitdomain.Pathspec{},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to git archive repo %s", config.IndexRepositoryName)
	}
	defer reader.Close()

	tr := tar.NewReader(reader)
	if err != nil {
		return err
	}

	didInsertNewCrates := false
	for {
		header, err := tr.Next()
		if err != nil {
			if err != io.EOF {
				return err
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

		var buf bytes.Buffer
		if _, err := io.CopyN(&buf, tr, header.Size); err != nil {
			return err
		}

		pkgs, err := parseCrateInformation(buf.String())
		if err != nil {
			return err
		}

		newCrates, newVersions, err := b.dependenciesSvc.InsertPackageRepoRefs(ctx, pkgs)
		if err != nil {
			return errors.Wrapf(err, "failed to insert Rust crate")
		}
		didInsertNewCrates = didInsertNewCrates || len(newCrates) != 0 || len(newVersions) != 0
	}

	if didInsertNewCrates {
		// We picked up new crates so we trigger a new sync for the RUSTPACKAGES code host.
		nextSync := time.Now()
		externalService.NextSyncAt = nextSync
		if err := b.extSvcStore.Upsert(ctx, externalService); err != nil {
			return err
		}
	}
	return nil
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
		return nil, errors.Wrapf(err, "failed to parse JSON config for Rust external service %s", rawConfig)
	}

	if err = jsoniter.Unmarshal(normalized, config); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal Rust external service config %s", rawConfig)
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
		return false, nil, errors.Wrapf(err, "failed to list Rust external service types")
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
func parseCrateInformation(contents string) ([]shared.MinimalPackageRepoRef, error) {
	crates := make(map[reposource.PackageName]*shared.MinimalPackageRepoRef, strings.Count(contents, "\n"))

	for _, line := range strings.Split(contents, "\n") {
		if line == "" {
			continue
		}

		type crateInfo struct {
			Name    string `json:"name"`
			Version string `json:"vers"`
		}
		var info crateInfo
		err := json.Unmarshal([]byte(line), &info)
		if err != nil {
			return nil, err
		}

		name := reposource.PackageName(info.Name)
		if crate, ok := crates[name]; ok {
			crate.Versions = append(crate.Versions, info.Version)
		} else {
			crates[name] = &shared.MinimalPackageRepoRef{
				Scheme:   shared.RustPackagesScheme,
				Name:     name,
				Versions: []string{info.Version},
			}
		}
	}

	result := make([]shared.MinimalPackageRepoRef, 0, len(crates))

	for _, crate := range crates {
		result = append(result, *crate)
	}

	return result, nil
}
