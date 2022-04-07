package codeintel

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
)

type (
	AutoIndexingService autoindexing.Service
	DependencyService   dependencies.Service
	DocumentsService    documents.Service
	PoliciesService     policies.Service
	SymbolsService      symbols.Service
	UploadsService      uploads.Service
)

var (
	GetAutoindexingService = autoindexing.GetService
	GetDependenciesService = dependencies.GetService
	GetDocumentsService    = documents.GetService
	GetPoliciesService     = policies.GetService
	GetSymbolsService      = symbols.GetService
	GetUploadsService      = uploads.GetService
)
