package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ResolvedDiagnostic struct {
	Dump       store.Dump
	Diagnostic lsifstore.Diagnostic
}

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (api *CodeIntelAPI) Diagnostics(ctx context.Context, prefix string, uploadID, limit, offset int) (_ []ResolvedDiagnostic, _ int, err error) {
	ctx, endObservation := api.operations.diagnostics.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("prefix", prefix),
		log.Int("uploadID", uploadID),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	dump, exists, err := api.dbStore.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, 0, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, 0, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(prefix, dump.Root)
	diagnostics, totalCount, err := api.lsifStore.Diagnostics(ctx, dump.ID, pathInBundle, offset, limit)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "bundleClient.Diagnostics")
	}

	return resolveDiagnosticsWithDump(dump, diagnostics), totalCount, nil
}

func resolveDiagnosticsWithDump(dump store.Dump, diagnostics []lsifstore.Diagnostic) []ResolvedDiagnostic {
	var resolvedDiagnostics []ResolvedDiagnostic
	for _, diagnostic := range diagnostics {
		diagnostic.Path = dump.Root + diagnostic.Path
		resolvedDiagnostics = append(resolvedDiagnostics, ResolvedDiagnostic{
			Dump:       dump,
			Diagnostic: diagnostic,
		})
	}

	return resolvedDiagnostics
}
