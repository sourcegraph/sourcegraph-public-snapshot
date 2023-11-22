// Package job contains the definitions and helpers for search jobs.
// This imports of this package should stay minimal so it can be referenced
// by other packages without pulling in a large set of transitive dependencies.
package job

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// Job is an interface shared by all individual search operations in the
// backend (e.g., text vs commit vs symbol search are represented as different
// jobs) as well as combinations over those searches (run a set in parallel,
// timeout). Calling Run on a job object runs a search.
type Job interface {
	Run(context.Context, RuntimeClients, streaming.Sender) (*search.Alert, error)

	// MapChildren recursively applies MapFunc to every child job of this job,
	// returning a copied job with the resulting set of children.
	MapChildren(MapFunc) Job

	Describer
}

// PartialJob is a partially constructed job that needs information only
// available at runtime to resolve a fully constructed job.
type PartialJob[T any] interface {
	// Resolve returns the fully constructed job using information that is only
	// available at runtime.
	Resolve(T) Job

	// MapChildren recursively applies MapFunc to every child job of this job,
	// returning a copied job with the resulting set of children.
	MapChildren(MapFunc) PartialJob[T]

	Describer
}

// Describer is in interface that allows a job to self-describe. It is shared
// by all jobs and partial jobs
type Describer interface {
	// Name is the name of the job
	Name() string

	// Children is the list of the job's children
	Children() []Describer

	// Fields is the set of fields that describe the job
	Attributes(Verbosity) []attribute.KeyValue
}

type Verbosity int

const (
	VerbosityNone  Verbosity = iota // no fields
	VerbosityBasic                  // essential fields
	VerbosityMax                    // all possible fields
)

type RuntimeClients struct {
	Logger                      log.Logger
	DB                          database.DB
	Zoekt                       zoekt.Streamer
	SearcherURLs                *endpoint.Map
	SearcherGRPCConnectionCache *defaults.ConnectionCache
	Gitserver                   gitserver.Client
}
