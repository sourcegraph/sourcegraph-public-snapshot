package background

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"

	"github.com/sourcegraph/batch-change-utils/env"
	"github.com/sourcegraph/batch-change-utils/overridable"
	"github.com/sourcegraph/batch-change-utils/yaml"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/schema"
)

type pendingBatchSpecHandler struct {
	s *store.Store
}

func (h *pendingBatchSpecHandler) HandlerFunc() dbworker.HandlerFunc {
	return func(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
		s := h.s.With(tx)
		log15.Info("processing pending batch spec", "record", record, "s", s)

		pbs := record.(pendingBatchSpecRecord).PendingBatchSpec
		spec, err := ParseBatchSpec([]byte(pbs.Spec))
		if err != nil {
			return errors.Wrap(err, "parsing batch spec")
		}

		// We need to resolve the repositories into workspaces, then create
		// BatchExecutionJob instances.

		return nil
	}
}

type pendingBatchSpecRecord struct {
	*btypes.PendingBatchSpec
}

var _ workerutil.Record = pendingBatchSpecRecord{}

func (r pendingBatchSpecRecord) RecordID() int {
	return int(r.ID)
}

func newPendingBatchSpecWorker(
	ctx context.Context,
	s *store.Store,
	metrics batchChangesMetrics,
) *workerutil.Worker {
	handler := &pendingBatchSpecHandler{s}
	workerStore := newPendingBatchSpecWorkerStore(s)
	options := workerutil.WorkerOptions{
		Name:        "batches_pending_batch_spec_worker",
		NumHandlers: 5,
		Interval:    5 * time.Second,
		Metrics:     metrics.pendingBatchSpecWorkerMetrics,
	}

	return dbworker.NewWorker(ctx, workerStore, handler.HandlerFunc(), options)
}

func newPendingBatchSpecWorkerResetter(s *store.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	workerStore := newPendingBatchSpecWorkerStore(s)
	options := dbworker.ResetterOptions{
		Name:     "batches_pending_batch_spec_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.pendingBatchSpecWorkerResetterMetrics,
	}

	return dbworker.NewResetter(workerStore, options)
}

func newPendingBatchSpecWorkerStore(s *store.Store) dbworkerstore.Store {
	return dbworkerstore.New(s.Handle(), dbworkerstore.Options{
		Name:              "batches_pending_batch_spec_store",
		TableName:         "pending_batch_specs",
		ColumnExpressions: store.PendingBatchSpecColumns,
		Scan: func(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
			pbs, exists, err := store.ScanFirstPendingBatchSpec(rows, err)
			return pendingBatchSpecRecord{pbs}, exists, err
		},
		OrderByExpression: sqlf.Sprintf("pending_batch_specs.updated_at ASC"),
		StalledMaxAge:     60 * time.Second,
		MaxNumResets:      60,
		RetryAfter:        5 * time.Second,
		MaxNumRetries:     60,
	})
}

type BatchSpec struct {
	Name              string                   `json:"name,omitempty" yaml:"name"`
	Description       string                   `json:"description,omitempty" yaml:"description"`
	On                []OnQueryOrRepository    `json:"on,omitempty" yaml:"on"`
	Workspaces        []WorkspaceConfiguration `json:"workspaces,omitempty"  yaml:"workspaces"`
	Steps             []Step                   `json:"steps,omitempty" yaml:"steps"`
	TransformChanges  *TransformChanges        `json:"transformChanges,omitempty" yaml:"transformChanges,omitempty"`
	ImportChangesets  []ImportChangeset        `json:"importChangesets,omitempty" yaml:"importChangesets"`
	ChangesetTemplate *ChangesetTemplate       `json:"changesetTemplate,omitempty" yaml:"changesetTemplate"`
}

type ChangesetTemplate struct {
	Title     string                       `json:"title,omitempty" yaml:"title"`
	Body      string                       `json:"body,omitempty" yaml:"body"`
	Branch    string                       `json:"branch,omitempty" yaml:"branch"`
	Commit    ExpandedGitCommitDescription `json:"commit,omitempty" yaml:"commit"`
	Published overridable.BoolOrString     `json:"published" yaml:"published"`
}

type GitCommitAuthor struct {
	Name  string `json:"name" yaml:"name"`
	Email string `json:"email" yaml:"email"`
}

type ExpandedGitCommitDescription struct {
	Message string           `json:"message,omitempty" yaml:"message"`
	Author  *GitCommitAuthor `json:"author,omitempty" yaml:"author"`
}

type ImportChangeset struct {
	Repository  string        `json:"repository" yaml:"repository"`
	ExternalIDs []interface{} `json:"externalIDs" yaml:"externalIDs"`
}

type WorkspaceConfiguration struct {
	RootAtLocationOf   string `json:"rootAtLocationOf,omitempty" yaml:"rootAtLocationOf"`
	In                 string `json:"in,omitempty" yaml:"in"`
	OnlyFetchWorkspace bool   `json:"onlyFetchWorkspace,omitempty" yaml:"onlyFetchWorkspace"`

	glob glob.Glob
}

type OnQueryOrRepository struct {
	RepositoriesMatchingQuery string `json:"repositoriesMatchingQuery,omitempty" yaml:"repositoriesMatchingQuery"`
	Repository                string `json:"repository,omitempty" yaml:"repository"`
	Branch                    string `json:"branch,omitempty" yaml:"branch"`
}

type Step struct {
	Run       string            `json:"run,omitempty" yaml:"run"`
	Container string            `json:"container,omitempty" yaml:"container"`
	Env       env.Environment   `json:"env,omitempty" yaml:"env"`
	Files     map[string]string `json:"files,omitempty" yaml:"files,omitempty"`
	Outputs   Outputs           `json:"outputs,omitempty" yaml:"outputs,omitempty"`

	If interface{} `json:"if,omitempty" yaml:"if,omitempty"`
}

type Outputs map[string]Output

type Output struct {
	Value  string `json:"value,omitempty" yaml:"value,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
}

type TransformChanges struct {
	Group []Group `json:"group,omitempty" yaml:"group"`
}

type Group struct {
	Directory  string `json:"directory,omitempty" yaml:"directory"`
	Branch     string `json:"branch,omitempty" yaml:"branch"`
	Repository string `json:"repository,omitempty" yaml:"repository"`
}

func ParseBatchSpec(data []byte) (*BatchSpec, error) {
	var spec BatchSpec
	if err := yaml.UnmarshalValidate(schema.BatchSpecSchemaJSON, data, &spec); err != nil {
		if multiErr, ok := err.(*multierror.Error); ok {
			var newMultiError *multierror.Error

			for _, e := range multiErr.Errors {
				// In case of `name` we try to make the error message more user-friendly.
				if strings.Contains(e.Error(), "name: Does not match pattern") {
					newMultiError = multierror.Append(newMultiError, fmt.Errorf("The batch change name can only contain word characters, dots and dashes. No whitespace or newlines allowed."))
				} else {
					newMultiError = multierror.Append(newMultiError, e)
				}
			}

			return nil, newMultiError.ErrorOrNil()
		}

		return nil, err
	}

	var errs *multierror.Error

	if len(spec.Steps) != 0 && spec.ChangesetTemplate == nil {
		errs = multierror.Append(errs, errors.New("batch spec includes steps but no changesetTemplate"))
	}

	return &spec, errs.ErrorOrNil()
}
