package cratesyncer

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type syncer struct {
	db            database.DB
	dbStore       *dbstore.Store
	extSvcStore   database.ExternalServiceStore
	indexEnqueuer *autoindexing.Service
	// gitserverClient *gitserver.Client
}

var _ goroutine.Handler = &syncer{}

type CrateInfo struct {
	Name    string `json:"name"`
	Version string `json:"vers"`
}

func (s *syncer) Handle(ctx context.Context) error {
	// TODO: Skip if not in our sourcegraph cloud instance
	// TODO: Don't always auto-index these repos

	fmt.Println("=============================== New Execution")

	kind := extsvc.KindRustPackages
	externalServices, err := s.extSvcStore.List(ctx, database.ExternalServicesListOptions{
		Kinds: []string{kind},
	})

	fmt.Println("EXTERNAL:", externalServices, err)
	if err != nil {
		return err
	}

	//  Skip if RUSTPACKAGES not enabled
	if len(externalServices) == 0 {
		return nil
	}

	repo := "github.com/rust-lang/crates.io-index"
	reader, err := git.ArchiveReader(ctx, s.db, nil, api.RepoName(repo), gitserver.ArchiveOptions{
		Treeish:   "HEAD",
		Format:    "tar",
		Pathspecs: []gitserver.Pathspec{},
	})
	if err != nil {
		fmt.Println("Archive Reader Err:", err)
		return err
	}
	defer reader.Close()

	tr := tar.NewReader(reader)
	if err != nil {
		return err
	}

	count := 0
	pkgs := []precise.Package{}
	for {
		header, err := tr.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}

			break
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

		if !strings.Contains(header.Name, "tokio") {
			continue
		}

		var buf bytes.Buffer
		if _, err := io.CopyN(&buf, tr, header.Size); err != nil {
			return err
		}

		contents := buf.String()
		for _, line := range strings.Split(contents, "\n") {
			if line == "" {
				continue
			}

			var info CrateInfo
			err = json.Unmarshal([]byte(line), &info)
			if err != nil {
				return err
			}

			pkg := precise.Package{
				Scheme:  dependencies.RustPackagesScheme,
				Name:    info.Name,
				Version: info.Version,
			}

			isNew, err := s.dbStore.InsertCloneableDependencyRepo(ctx, pkg)
			if err != nil {
				return err
			}

			fmt.Println("IS NEW:", isNew)
			pkgs = append(pkgs, pkg)
		}

		if count > 5 {
			break
		}

		count = count + 1
	}

	nextSync := time.Now()
	for _, externalService := range externalServices {
		externalService.NextSyncAt = nextSync
		if err := s.extSvcStore.Upsert(ctx, externalService); err != nil {
			return err
		}
	}

	for _, pkg := range pkgs {
		fmt.Println("PKG:", pkg)

		if err := s.indexEnqueuer.QueueIndexesForPackage(ctx, pkg); err != nil {
			return nil
		}
	}

	fmt.Println("====> All Done")
	return nil
}

func NewCratesSyncer(db database.DB, indexEnqueuer *autoindexing.Service) goroutine.BackgroundRoutine {
	observationContext := &observation.Context{
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.NewRegistry(),
	}
	extSvcStore := database.NewDB(db).ExternalServices()

	hour := time.Hour
	duration := hour * 12
	return goroutine.NewPeriodicGoroutine(context.Background(), duration, &syncer{
		db:            db,
		dbStore:       dbstore.NewWithDB(db, observationContext),
		extSvcStore:   extSvcStore,
		indexEnqueuer: indexEnqueuer,
	})
}
