package cratesyncer

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	dbtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type syncer struct {
	db                    database.DB
	dbStore               *dbstore.Store
	externalServicesStore database.ExternalServiceStore
	gitClient             gitserver.Client
	interval              time.Duration
}

var _ goroutine.Handler = &syncer{}

func (s *syncer) Handle(ctx context.Context) error {

	exists, externalService, err := singleRustExternalService(ctx, s.externalServicesStore)
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
	update, err := s.gitClient.RequestRepoUpdate(ctx, repoName, s.interval)
	if err != nil {
		return err
	}
	if update != nil && update.Error != "" {
		return errors.Newf("failed to update repo %s, error %s", repoName, update.Error)
	}
	reader, err := s.gitClient.ArchiveReader(
		ctx,
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
		for _, pkg := range pkgs {
			// TODO: batch insert packages instead of one name+version combination at a time https://github.com/sourcegraph/sourcegraph/issues/37691
			isNew, err := s.dbStore.InsertCloneableDependencyRepo(ctx, pkg)
			if err != nil {
				return errors.Wrapf(err, "Failed to insert Rust crate %v", pkg)
			}
			didInsertNewCrates = didInsertNewCrates || isNew
		}
	}

	if didInsertNewCrates {
		// We picked up new crates so we trigger a new sync for the RUSTPACKAGES code host.
		nextSync := time.Now()
		externalService.NextSyncAt = nextSync
		if err := s.externalServicesStore.Upsert(ctx, externalService); err != nil {
			return err
		}
	}
	return nil
}

func NewCratesSyncer(db database.DB) goroutine.BackgroundRoutine {
	observationContext := &observation.Context{
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.NewRegistry(),
	}
	extSvcStore := db.ExternalServices()

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

	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &syncer{
		db:                    db,
		dbStore:               dbstore.NewWithDB(db, observationContext),
		externalServicesStore: extSvcStore,
		gitClient:             gitserver.NewClient(db),
		interval:              interval,
	})
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
func singleRustExternalService(ctx context.Context, store database.ExternalServiceStore) (exists bool, externalService *dbtypes.ExternalService, err error) {
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
func parseCrateInformation(contents string) ([]precise.Package, error) {
	var result []precise.Package
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

		pkg := precise.Package{
			Scheme:  dependencies.RustPackagesScheme,
			Name:    info.Name,
			Version: info.Version,
		}
		result = append(result, pkg)

	}
	return result, nil
}
