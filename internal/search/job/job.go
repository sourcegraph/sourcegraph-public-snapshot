// Package job contains the definitions and helpers for search jobs.
// This imports of this package should stay minimal so it can be referenced
// by other packages without pulling in a large set of transitive dependencies.
package job

import (
	"context"

	"github.com/google/zoekt"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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
	// Partial returns the partially constructed job. This interface allows us
	// to inspect the state of a parameterized job before resolving it, which
	// happens at runtime.
	Partial() Job

	// Resolve returns the fully constructed job using information that is only
	// available at runtime.
	Resolve(T) Job

	MapChildren(MapFunc) PartialJob[T]
	Describer
}

type Describer interface {
	// Name is the name of the job
	Name() string

	// Children is the list of the job's children
	Children() []Describer

	// Fields is the set of fields that describe the job
	Fields(Verbosity) []otlog.Field
}

type Verbosity int

const (
	VerbosityNone  Verbosity = iota // no fields
	VerbosityBasic                  // essential fields
	VerbosityMax                    // all possible fields
)

type RuntimeClients struct {
	Logger       log.Logger
	DB           database.DB
	Zoekt        zoekt.Streamer
	SearcherURLs *endpoint.Map
	Gitserver    gitserver.Client
}

type MapFunc func(Job) Job

// Map applies fn to every job in tree recursively, returning a new job.
// The provided function should return a copied job with the mutations
// applied rather than mutating the job in-place.
func Map(j Job, fn MapFunc) Job {
	j = j.MapChildren(fn)
	return fn(j)
}
