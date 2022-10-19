package graphqlbackend

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

const oneReleaseCycle = 35 * 24 * time.Hour

var insiderBuildRegex = regexp.MustCompile(`^[\w-]+_(\d{4}-\d{2}-\d{2})_\w+`)

func NewExecutorResolver(executor types.Executor) *ExecutorResolver {
	return &ExecutorResolver{executor: executor}
}

type ExecutorResolver struct {
	executor types.Executor
}

func (e *ExecutorResolver) ID() graphql.ID {
	return relay.MarshalID("Executor", (int64(e.executor.ID)))
}
func (e *ExecutorResolver) Hostname() string  { return e.executor.Hostname }
func (e *ExecutorResolver) QueueName() string { return e.executor.QueueName }
func (e *ExecutorResolver) Active() bool {
	// TODO: Read the value of the executor worker heartbeat interval in here.
	heartbeatInterval := 5 * time.Second
	return time.Since(e.executor.LastSeenAt) <= 3*heartbeatInterval
}
func (e *ExecutorResolver) Os() string              { return e.executor.OS }
func (e *ExecutorResolver) Architecture() string    { return e.executor.Architecture }
func (e *ExecutorResolver) DockerVersion() string   { return e.executor.DockerVersion }
func (e *ExecutorResolver) ExecutorVersion() string { return e.executor.ExecutorVersion }
func (e *ExecutorResolver) GitVersion() string      { return e.executor.GitVersion }
func (e *ExecutorResolver) IgniteVersion() string   { return e.executor.IgniteVersion }
func (e *ExecutorResolver) SrcCliVersion() string   { return e.executor.SrcCliVersion }
func (e *ExecutorResolver) FirstSeenAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: e.executor.FirstSeenAt}
}
func (e *ExecutorResolver) LastSeenAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: e.executor.LastSeenAt}
}

func (e *ExecutorResolver) Compatibility() (*string, error) {
	ev := e.executor.ExecutorVersion
	if !e.Active() {
		return nil, nil
	}
	return calculateExecutorCompatibility(ev)
}

func calculateExecutorCompatibility(ev string) (*string, error) {
	var compatibility ExecutorCompatibility = ExecutorCompatibilityUpToDate
	sv := version.Version()

	isExecutorDev := ev != "" && version.IsDev(ev)
	isSgDev := sv != "" && version.IsDev(sv)

	if isSgDev || isExecutorDev {
		return nil, nil
	}

	evm := insiderBuildRegex.FindStringSubmatch(ev)
	svm := insiderBuildRegex.FindStringSubmatch(sv)

	// check for version mismatch
	if len(evm) > 1 && len(svm) <= 1 {
		// this means that the executor is an insider version while the Sourcegraph
		// instance is not.
		return nil, nil
	}

	if len(evm) <= 1 && len(svm) > 1 {
		// this means that the Sourcegraph instance is an insider version while the
		// executor is not.
		return nil, nil
	}

	if len(evm) > 1 && len(svm) > 1 {
		layout := "2006-01-02"

		st, err := time.Parse(layout, svm[1])
		if err != nil {
			return nil, err
		}

		et, err := time.Parse(layout, evm[1])
		if err != nil {
			return nil, err
		}

		hst := st.Add(oneReleaseCycle)
		lst := st.Add(-1 * oneReleaseCycle)

		if et.After(hst) {
			// We check if the executor build date is after a release cycle + sourcegraph build date.
			// if this is true then we assume the executor's version is ahead.
			compatibility = ExecutorCompatibilityVersionAhead
		} else if et.Before(lst) {
			// if the executor date is a release cycle behind the current build date of the Sourcegraph
			// instance then we assume that the executor is outdated.
			compatibility = ExecutorCompatibilityOutdated
		}

		return compatibility.ToGraphQL(), nil
	}

	s, err := semver.NewVersion(sv)
	if err != nil {
		return nil, err
	}

	e, err := semver.NewVersion(ev)
	if err != nil {
		return nil, err
	}

	// it's okay for an executor to be one minor version behind or ahead of the sourcegraph version.
	iev := e.IncMinor()

	isv := s.IncMinor()

	if s.GreaterThan(&iev) {
		compatibility = ExecutorCompatibilityOutdated
	} else if isv.LessThan(e) {
		compatibility = ExecutorCompatibilityVersionAhead
	}

	return compatibility.ToGraphQL(), nil
}

func marshalExecutorSecretID(id int64) graphql.ID {
	return relay.MarshalID("ExecutorSecret", id)
}

func unmarshalExecutorSecretID(gqlID graphql.ID) (id int64, err error) {
	err = relay.UnmarshalSpec(gqlID, &id)
	return
}

type ExecutorSecretResolver struct {
	db     database.DB
	secret *database.ExecutorSecret
}

func (e *ExecutorSecretResolver) ID() graphql.ID {
	return marshalExecutorSecretID(e.secret.ID)
}
func (e *ExecutorSecretResolver) Key() string   { return e.secret.Key }
func (e *ExecutorSecretResolver) Scope() string { return strings.ToUpper(e.secret.Scope) }
func (e *ExecutorSecretResolver) Namespace(ctx context.Context) (*NamespaceResolver, error) {
	if e.secret.NamespaceUserID != 0 {
		n, err := UserByIDInt32(ctx, e.db, e.secret.NamespaceUserID)
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}

	if e.secret.NamespaceOrgID != 0 {
		n, err := OrgByIDInt32(ctx, e.db, e.secret.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}

	return nil, nil
}
func (e *ExecutorSecretResolver) Creator(ctx context.Context) (*UserResolver, error) {
	// User has been deleted.
	if e.secret.CreatorID == 0 {
		return nil, nil
	}

	return nil, nil
}
func (e *ExecutorSecretResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: e.secret.CreatedAt}
}
func (e *ExecutorSecretResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: e.secret.UpdatedAt}
}

type ExecutorSecretAccessLogListArgs struct {
	First int32
	After *string
}

func (r *ExecutorSecretResolver) AccessLogs(args ExecutorSecretAccessLogListArgs) (*executorSecretAccessLogConnectionResolver, error) {
	limit := &database.LimitOffset{Limit: int(args.First)}
	if args.After != nil {
		offset, err := graphqlutil.DecodeIntCursor(args.After)
		if err != nil {
			return nil, err
		}
		limit.Offset = offset
	}

	return &executorSecretAccessLogConnectionResolver{
		opts: database.ExecutorSecretAccessLogsListOpts{
			LimitOffset:      limit,
			ExecutorSecretID: r.secret.ID,
		},
		db: r.db,
	}, nil
}

type executorSecretAccessLogConnectionResolver struct {
	db   database.DB
	opts database.ExecutorSecretAccessLogsListOpts

	computeOnce sync.Once
	logs        []*database.ExecutorSecretAccessLog
	next        int
	err         error
}

func (r *executorSecretAccessLogConnectionResolver) Nodes(ctx context.Context) ([]*executorSecretAccessLogResolver, error) {
	logs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*executorSecretAccessLogResolver, 0, len(logs))
	for _, log := range logs {
		resolvers = append(resolvers, &executorSecretAccessLogResolver{db: r.db, log: log})
	}

	return resolvers, nil
}

func (r *executorSecretAccessLogConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	totalCount, err := r.db.ExecutorSecretAccessLogs().Count(ctx, r.opts)
	return int32(totalCount), err
}

func (r *executorSecretAccessLogConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != 0 {
		n := int32(next)
		return graphqlutil.EncodeIntCursor(&n), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *executorSecretAccessLogConnectionResolver) compute(ctx context.Context) (_ []*database.ExecutorSecretAccessLog, next int, err error) {
	r.computeOnce.Do(func() {
		r.logs, r.next, r.err = r.db.ExecutorSecretAccessLogs().List(ctx, r.opts)
	})
	return r.logs, r.next, r.err
}

func marshalExecutorSecretAccessLogID(id int64) graphql.ID {
	return relay.MarshalID("ExecutorSecretAccessLog", id)
}

func unmarshalExecutorSecretAccessLogID(gqlID graphql.ID) (id int64, err error) {
	err = relay.UnmarshalSpec(gqlID, &id)
	return
}

type executorSecretAccessLogResolver struct {
	db  database.DB
	log *database.ExecutorSecretAccessLog
}

func (r *executorSecretAccessLogResolver) ID() graphql.ID {
	return marshalExecutorSecretAccessLogID(r.log.ID)
}

func (r *executorSecretAccessLogResolver) ExecutorSecret(ctx context.Context) (*ExecutorSecretResolver, error) {
	return executorSecretByID(ctx, r.db, marshalExecutorSecretID(r.log.ExecutorSecretID))
}

func (r *executorSecretAccessLogResolver) User(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.log.UserID)
}

func (r *executorSecretAccessLogResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.log.CreatedAt}
}
