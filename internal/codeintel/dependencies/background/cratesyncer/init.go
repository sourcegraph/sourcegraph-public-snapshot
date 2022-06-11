package cratesyncer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type syncer struct {
	dbStore       *dbstore.Store
	extSvcStore   database.ExternalServiceStore
	indexEnqueuer *autoindexing.Service
}

type CrateInfo struct {
	Name    string `json:"name"`
	Version string `json:"vers"`
}

func (s *syncer) Handle(ctx context.Context) error {
	fmt.Println("===============================")
	fmt.Println(" ... Doing something ...", s.dbStore)

	kind := extsvc.KindRustPackages
	externalServices, err := s.extSvcStore.List(ctx, database.ExternalServicesListOptions{
		Kinds: []string{kind},
	})

	fmt.Println("EXTERNAL:", externalServices, err)
	if err != nil {
		return err
	}

	pkgs := []precise.Package{}
	err = filepath.Walk("/home/tjdevries/git/crates.io-index/to/ki/",
		func(path string, fileinfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if fileinfo.IsDir() {
				return nil
			}

			fmt.Println(path, fileinfo.Size())
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			contents := string(data)
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

				s.dbStore.InsertCloneableDependencyRepo(ctx, pkg)
				pkgs = append(pkgs, pkg)
			}

			return nil
		})

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

	fmt.Println("===============================")
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
		dbStore:       dbstore.NewWithDB(db, observationContext),
		extSvcStore:   extSvcStore,
		indexEnqueuer: indexEnqueuer,
	})
}
