package resolvers

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"golang.org/x/sync/semaphore"
)

const actionExecutionIDKind = "ActionExecution"

func marshalActionExecutionID(id int64) graphql.ID {
	return relay.MarshalID(actionExecutionIDKind, id)
}

func unmarshalActionExecutionID(id graphql.ID) (actionExecutionID int64, err error) {
	err = relay.UnmarshalSpec(id, &actionExecutionID)
	return
}

type actionExecutionResolver struct {
	store *ee.Store

	actionExecution campaigns.ActionExecution

	// todo: use this for passing down the action when a parent resolver was an action resolver
	action *campaigns.Action

	// pass this when the jobs are already known:
	actionJobs *[]*campaigns.ActionJob
}

func (r *Resolver) ActionExecutionByID(ctx context.Context, id graphql.ID) (graphqlbackend.ActionExecutionResolver, error) {
	// todo: permissions
	dbId, err := unmarshalActionExecutionID(id)
	if err != nil {
		return nil, err
	}

	actionExecution, err := r.store.ActionExecutionByID(ctx, ee.ActionExecutionByIDOpts{
		ID: dbId,
	})
	if err != nil {
		return nil, err
	}
	if actionExecution.ID == 0 {
		return nil, nil
	}

	return &actionExecutionResolver{store: r.store, actionExecution: *actionExecution}, nil
}

func (r *actionExecutionResolver) ID() graphql.ID {
	return marshalActionExecutionID(r.actionExecution.ID)
}

func (r *actionExecutionResolver) Action(ctx context.Context) (graphqlbackend.ActionResolver, error) {
	if r.action != nil {
		return &actionResolver{store: r.store, action: *r.action}, nil
	}
	action, err := r.store.ActionByID(ctx, ee.ActionByIDOpts{ID: r.actionExecution.ActionID})
	if err != nil {
		return nil, err
	}
	return &actionResolver{store: r.store, action: *action}, nil
}

func (r *actionExecutionResolver) InvokationReason() campaigns.ActionExecutionInvokationReason {
	return r.actionExecution.InvokationReason
}

func (r *actionExecutionResolver) Definition() graphqlbackend.ActionDefinitionResolver {
	return &actionDefinitionResolver{steps: r.actionExecution.Steps, envStr: *r.actionExecution.EnvStr}
}

// todo:
func (r *actionExecutionResolver) ActionWorkspace() *graphqlbackend.GitTreeEntryResolver {
	return nil
}

func (r *actionExecutionResolver) Jobs() graphqlbackend.ActionJobConnectionResolver {
	return &actionJobConnectionResolver{store: r.store, actionExecution: &r.actionExecution, knownJobs: r.actionJobs}
}

func (r *actionExecutionResolver) Status(ctx context.Context) (*campaigns.BackgroundProcessStatus, error) {
	status, err := r.store.ActionExecutionStatus(ctx, ee.ActionExecutionStatusOpts{ExecutionID: r.actionExecution.ID})
	if err != nil {
		return nil, err
	}

	var processState campaigns.BackgroundProcessState = campaigns.BackgroundProcessStateProcessing
	// canceled has higher precendence than errored
	if status.Canceled == true {
		processState = campaigns.BackgroundProcessStateCanceled
	} else if status.Pending == 0 {
		// todo: currently, still running has precedence over errored, revisit if that's useful
		if status.Errored == true {
			processState = campaigns.BackgroundProcessStateErrored
		} else {
			processState = campaigns.BackgroundProcessStateCompleted
		}
	}

	return &campaigns.BackgroundProcessStatus{
		Canceled:     status.Canceled,
		Total:        int32(status.Total),
		Completed:    int32(status.Total - status.Pending),
		Pending:      int32(status.Pending),
		ProcessState: processState,
		// todo: we don't have the errors as single line strings
		ProcessErrors: []string{},
	}, nil
}

func (r *actionExecutionResolver) ExecutionStart() *graphqlbackend.DateTime {
	if r.actionExecution.ExecutionStart.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.actionExecution.ExecutionStart}
}

func (r *actionExecutionResolver) ExecutionEnd() *graphqlbackend.DateTime {
	if r.actionExecution.ExecutionEnd.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.actionExecution.ExecutionEnd}
}

func (r *actionExecutionResolver) PatchSet(ctx context.Context) (graphqlbackend.PatchSetResolver, error) {
	if r.actionExecution.PatchSetID == nil {
		return nil, nil
	}
	patchSet, err := r.store.GetPatchSet(ctx, ee.GetPatchSetOpts{ID: *r.actionExecution.PatchSetID})
	if err != nil {
		return nil, err
	}

	return &patchSetResolver{store: r.store, patchSet: patchSet}, nil
}

func createActionExecutionForAction(ctx context.Context, store *ee.Store, action *campaigns.Action, invokationReason campaigns.ActionExecutionInvokationReason) (*campaigns.ActionExecution, []*campaigns.ActionJob, error) {
	scopeQuery, err := scopeQueryForSteps(action.Steps)
	if err != nil {
		return nil, nil, err
	}
	repos, err := findRepos(ctx, scopeQuery)
	if err != nil {
		return nil, nil, err
	}
	if len(repos) == 0 {
		return nil, nil, errors.New("Cannot create execution for action that yields 0 repositories")
	}

	tx, err := store.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Done(&err)

	actionExecution, err := tx.CreateActionExecution(ctx, ee.CreateActionExecutionOpts{
		InvokationReason: invokationReason,
		Steps:            action.Steps,
		EnvStr:           action.EnvStr,
		ActionID:         action.ID,
	})
	if err != nil {
		return nil, nil, err
	}
	actionJobs := make([]*campaigns.ActionJob, len(repos))
	for i, repo := range repos {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(graphql.ID(repo.ID))
		if err != nil {
			return nil, nil, err
		}
		// todo: caching
		actionJob, err := tx.CreateActionJob(ctx, ee.CreateActionJobOpts{
			ExecutionID:   actionExecution.ID,
			RepositoryID:  int64(repoID),
			BaseRevision:  repo.Rev,
			BaseReference: repo.Ref,
		})
		if err != nil {
			return nil, nil, err
		}
		actionJobs[i] = actionJob
	}
	return actionExecution, actionJobs, nil
}

func scopeQueryForSteps(actionFile string) (string, error) {
	var action struct {
		ScopeQuery string `json:"scopeQuery,omitempty"`
	}
	if err := jsonc.Unmarshal(string(actionFile), &action); err != nil {
		return "", errors.Wrap(err, "invalid JSON action file")
	}
	return action.ScopeQuery, nil
}

type actionRepo struct {
	ID  string
	Rev string
	Ref string
}

func findRepos(ctx context.Context, scopeQuery string) ([]actionRepo, error) {
	hasCount, err := regexp.MatchString(`count:\d+`, scopeQuery)
	if err != nil {
		return nil, err
	}

	if !hasCount {
		scopeQuery = scopeQuery + " count:999999"
	}
	search, err := graphqlbackend.NewSearchImplementer(&graphqlbackend.SearchArgs{
		Version: "V2",
		Query:   scopeQuery,
	})

	resultsResolver, err := search.Results(ctx)
	if err != nil {
		return nil, err
	}
	// unique map of all repos that matched the scope query
	repoMap := make(map[string]*graphqlbackend.RepositoryResolver)
	for _, _repo := range resultsResolver.Results() {
		repo, ok := _repo.ToRepository()
		if !ok {
			fm, ok := _repo.ToFileMatch()
			if !ok {
				return []actionRepo{}, errors.New("no valid search result")
			}
			repo = fm.Repository()
		}
		repoMap[repo.Name()] = repo
	}
	var wg sync.WaitGroup
	var repoMutex sync.Mutex
	sem := semaphore.NewWeighted(16)
	repos := make([]actionRepo, 0)
	for _, r := range repoMap {
		wg.Add(1)
		sem.Acquire(ctx, 1)
		go func(repo *graphqlbackend.RepositoryResolver) {
			defer wg.Done()
			defer sem.Release(1)
			defaultBranch, err := repo.DefaultBranch(ctx)
			if err != nil || defaultBranch == nil {
				fmt.Printf("# Skipping repository %s because we couldn't determine default branch.", repo.Name())
				return
			}
			target := defaultBranch.Target()
			oid, err := target.OID(ctx)
			if err != nil {
				fmt.Printf("# Skipping repository %s because we couldn't determine OID.", repo.Name())
				return
			}
			repoMutex.Lock()
			repos = append(repos, actionRepo{
				ID:  string(repo.ID()),
				Rev: string(oid),
				Ref: string(defaultBranch.Name()),
			})
			repoMutex.Unlock()
		}(r)
	}
	wg.Wait()
	return repos, nil
}
