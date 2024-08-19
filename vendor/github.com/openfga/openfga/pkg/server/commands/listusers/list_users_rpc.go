package listusers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	serverconfig "github.com/openfga/openfga/internal/server/config"

	"github.com/openfga/openfga/pkg/telemetry"

	"github.com/openfga/openfga/pkg/logger"

	"github.com/openfga/openfga/pkg/storage/storagewrappers"

	"github.com/openfga/openfga/internal/condition"
	"github.com/openfga/openfga/internal/condition/eval"
	"github.com/openfga/openfga/internal/graph"
	"github.com/openfga/openfga/internal/validation"
	"github.com/openfga/openfga/pkg/storage"
	"github.com/openfga/openfga/pkg/tuple"
	"github.com/openfga/openfga/pkg/typesystem"
)

var tracer = otel.Tracer("openfga/pkg/server/commands/list_users")

type listUsersQuery struct {
	logger                  logger.Logger
	ds                      storage.RelationshipTupleReader
	typesystemResolver      typesystem.TypesystemResolverFunc
	resolveNodeBreadthLimit uint32
	resolveNodeLimit        uint32
	maxResults              uint32
	maxConcurrentReads      uint32
	deadline                time.Duration
}

type expandResponse struct {
	hasCycle bool
	err      error
}

// userRelationshipStatus represents the status of a relationship that a given user/subject has with respect to a specific relation.
//
// A user/subject either does or does not have a relationship, which represents that
// they either explicitly do have a relationship or explicitly do not.
type userRelationshipStatus int

const (
	HasRelationship userRelationshipStatus = iota
	NoRelationship
)

type foundUser struct {
	user          *openfgav1.User
	excludedUsers []*openfgav1.User

	// relationshipStatus indicates whether the user explicitly does or does not have
	// a specific relationship with respect to the relation being expanded. It almost
	// exclusively applies to behavior stemming from exclusion rewrite rules.
	//
	// As users/subjects are being expanded we propagate the relationship status with
	// respect to the relation being evaluated so that we can handle subjects which
	// have been explicitly excluded from a relationship and where that relation is
	// contained under the subtracted branch of another exclusion. This allows us to
	// buble up the subject from the subtracted branch of the exclusion.
	relationshipStatus userRelationshipStatus
}

type ListUsersQueryOption func(l *listUsersQuery)

func WithListUsersQueryLogger(l logger.Logger) ListUsersQueryOption {
	return func(d *listUsersQuery) {
		d.logger = l
	}
}

// WithListUsersMaxResults see server.WithListUsersMaxResults.
func WithListUsersMaxResults(max uint32) ListUsersQueryOption {
	return func(d *listUsersQuery) {
		d.maxResults = max
	}
}

// WithListUsersDeadline see server.WithListUsersDeadline.
func WithListUsersDeadline(t time.Duration) ListUsersQueryOption {
	return func(d *listUsersQuery) {
		d.deadline = t
	}
}

// WithResolveNodeLimit see server.WithResolveNodeLimit.
func WithResolveNodeLimit(limit uint32) ListUsersQueryOption {
	return func(d *listUsersQuery) {
		d.resolveNodeLimit = limit
	}
}

// WithResolveNodeBreadthLimit see server.WithResolveNodeBreadthLimit.
func WithResolveNodeBreadthLimit(limit uint32) ListUsersQueryOption {
	return func(d *listUsersQuery) {
		d.resolveNodeBreadthLimit = limit
	}
}

// WithListUsersMaxConcurrentReads see server.WithMaxConcurrentReadsForListUsers.
func WithListUsersMaxConcurrentReads(limit uint32) ListUsersQueryOption {
	return func(d *listUsersQuery) {
		d.maxConcurrentReads = limit
	}
}

// NewListUsersQuery is not meant to be shared.
func NewListUsersQuery(ds storage.RelationshipTupleReader, opts ...ListUsersQueryOption) *listUsersQuery {
	l := &listUsersQuery{
		logger: logger.NewNoopLogger(),
		ds:     ds,
		typesystemResolver: func(ctx context.Context, storeID, modelID string) (*typesystem.TypeSystem, error) {
			typesys, exists := typesystem.TypesystemFromContext(ctx)
			if !exists {
				return nil, fmt.Errorf("typesystem not provided in context")
			}

			return typesys, nil
		},
		resolveNodeBreadthLimit: serverconfig.DefaultResolveNodeBreadthLimit,
		resolveNodeLimit:        serverconfig.DefaultResolveNodeLimit,
		deadline:                serverconfig.DefaultListUsersDeadline,
		maxResults:              serverconfig.DefaultListUsersMaxResults,
		maxConcurrentReads:      serverconfig.DefaultMaxConcurrentReadsForListUsers,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// ListUsers assumes that the typesystem is in the context and that the request is valid.
func (l *listUsersQuery) ListUsers(
	ctx context.Context,
	req *openfgav1.ListUsersRequest,
) (*listUsersResponse, error) {
	ctx, span := tracer.Start(ctx, "ListUsers")
	defer span.End()

	cancellableCtx, cancelCtx := context.WithCancel(ctx)
	if l.deadline != 0 {
		cancellableCtx, cancelCtx = context.WithTimeout(cancellableCtx, l.deadline)
		defer cancelCtx()
	}
	defer cancelCtx()

	l.ds = storagewrappers.NewCombinedTupleReader(
		storagewrappers.NewBoundedConcurrencyTupleReader(l.ds, l.maxConcurrentReads),
		req.GetContextualTuples(),
	)
	typesys, ok := typesystem.TypesystemFromContext(cancellableCtx)
	if !ok {
		return nil, fmt.Errorf("typesystem missing in context")
	}

	userFilter := req.GetUserFilters()[0]
	isReflexiveUserset := userFilter.GetType() == req.GetObject().GetType() && userFilter.GetRelation() == req.GetRelation()

	if !isReflexiveUserset {
		hasPossibleEdges, err := doesHavePossibleEdges(typesys, req)
		if err != nil {
			return nil, err
		}
		if !hasPossibleEdges {
			span.SetAttributes(attribute.Bool("no_possible_edges", true))
			return &listUsersResponse{
				Users: []*openfgav1.User{},
				Metadata: listUsersResponseMetadata{
					DatastoreQueryCount: 0,
					DispatchCounter:     new(atomic.Uint32),
				},
			}, nil
		}
	}

	datastoreQueryCount := atomic.Uint32{}
	dispatchCount := atomic.Uint32{}

	foundUsersCh := l.buildResultsChannel()
	expandErrCh := make(chan error, 1)

	foundUsersUnique := make(map[tuple.UserString]foundUser, 1000)
	excludedUsersUnique := make(map[tuple.UserString]struct{}, 1000)

	doneWithFoundUsersCh := make(chan struct{}, 1)
	go func() {
		for foundUser := range foundUsersCh {
			foundUsersUnique[tuple.UserProtoToString(foundUser.user)] = foundUser
			for _, exception := range foundUser.excludedUsers {
				excludedUsersUnique[tuple.UserProtoToString(exception)] = struct{}{}
			}

			if l.maxResults > 0 {
				if uint32(len(foundUsersUnique)) >= l.maxResults {
					span.SetAttributes(attribute.Bool("max_results_found", true))
					break
				}
			}
		}

		doneWithFoundUsersCh <- struct{}{}
	}()

	go func() {
		internalRequest := fromListUsersRequest(req, &datastoreQueryCount, &dispatchCount)
		resp := l.expand(cancellableCtx, internalRequest, foundUsersCh)
		// first send error and then close results channel, to ensure that error takes precedence
		if resp.err != nil {
			expandErrCh <- resp.err
		}
		close(foundUsersCh)
	}()

	select {
	// Note: if all cases can proceed, one will be selected at random
	case err := <-expandErrCh:
		telemetry.TraceError(span, err)
		return nil, err
	case <-doneWithFoundUsersCh:
		break
	case <-cancellableCtx.Done():
		// to avoid a race on the 'foundUsersUnique' map below, wait for the range over the channel to close
		<-doneWithFoundUsersCh
		break
	}

	cancelCtx()

	foundUsers := make([]*openfgav1.User, 0, len(foundUsersUnique))
	for foundUserKey, foundUser := range foundUsersUnique {
		if foundUser.relationshipStatus == NoRelationship {
			continue
		}

		foundUsers = append(foundUsers, tuple.StringToUserProto(foundUserKey))
	}

	var excludedUsers []*openfgav1.ObjectOrUserset
	if len(foundUsers) > 0 {
		excludedUsers = make([]*openfgav1.ObjectOrUserset, 0, len(excludedUsersUnique))
		for foundExcludedUser := range excludedUsersUnique {
			excludedUsers = append(excludedUsers, tuple.StringToObjectOrUserset(foundExcludedUser))
		}
	}

	span.SetAttributes(attribute.Int("result_count", len(foundUsers)))

	return &listUsersResponse{
		Users:         foundUsers,
		ExcludedUsers: excludedUsers,
		Metadata: listUsersResponseMetadata{
			DatastoreQueryCount: datastoreQueryCount.Load(),
			DispatchCounter:     &dispatchCount,
		},
	}, nil
}

func doesHavePossibleEdges(typesys *typesystem.TypeSystem, req *openfgav1.ListUsersRequest) (bool, error) {
	g := graph.New(typesys)

	userFilters := req.GetUserFilters()

	source := typesystem.DirectRelationReference(userFilters[0].GetType(), userFilters[0].GetRelation())
	target := typesystem.DirectRelationReference(req.GetObject().GetType(), req.GetRelation())

	edges, err := g.GetPrunedRelationshipEdges(target, source)
	if err != nil {
		return false, err
	}

	return len(edges) > 0, err
}

func (l *listUsersQuery) dispatch(
	ctx context.Context,
	req *internalListUsersRequest,
	foundUsersChan chan<- foundUser,
) expandResponse {
	req.dispatchCount.Add(1)
	return l.expand(ctx, req, foundUsersChan)
}

func (l *listUsersQuery) expand(
	ctx context.Context,
	req *internalListUsersRequest,
	foundUsersChan chan<- foundUser,
) expandResponse {
	ctx, span := tracer.Start(ctx, "expand")
	defer span.End()
	span.SetAttributes(attribute.Int("depth", int(req.depth)))
	if req.depth >= l.resolveNodeLimit {
		return expandResponse{
			err: graph.ErrResolutionDepthExceeded,
		}
	}
	req.depth++

	if enteredCycle(req) {
		span.SetAttributes(attribute.Bool("cycle_detected", true))
		return expandResponse{
			hasCycle: true,
		}
	}

	reqObjectType := req.GetObject().GetType()
	reqObjectID := req.GetObject().GetId()
	reqRelation := req.GetRelation()

	for _, userFilter := range req.GetUserFilters() {
		if reqObjectType == userFilter.GetType() && reqRelation == userFilter.GetRelation() {
			trySendResult(ctx, foundUser{
				user: &openfgav1.User{
					User: &openfgav1.User_Userset{
						Userset: &openfgav1.UsersetUser{
							Type:     reqObjectType,
							Id:       reqObjectID,
							Relation: reqRelation,
						},
					},
				},
			}, foundUsersChan)
		}
	}

	typesys, err := l.typesystemResolver(ctx, req.GetStoreId(), req.GetAuthorizationModelId())
	if err != nil {
		return expandResponse{
			err: err,
		}
	}

	targetObjectType := req.GetObject().GetType()
	targetRelation := req.GetRelation()

	relation, err := typesys.GetRelation(targetObjectType, targetRelation)
	if err != nil {
		return expandResponse{
			err: err,
		}
	}

	relationRewrite := relation.GetRewrite()
	resp := l.expandRewrite(ctx, req, relationRewrite, foundUsersChan)
	if resp.err != nil {
		telemetry.TraceError(span, resp.err)
	}
	return resp
}

func (l *listUsersQuery) expandRewrite(
	ctx context.Context,
	req *internalListUsersRequest,
	rewrite *openfgav1.Userset,
	foundUsersChan chan<- foundUser,
) expandResponse {
	ctx, span := tracer.Start(ctx, "expandRewrite")
	defer span.End()

	var resp expandResponse
	switch rewrite := rewrite.GetUserset().(type) {
	case *openfgav1.Userset_This:
		resp = l.expandDirect(ctx, req, foundUsersChan)
	case *openfgav1.Userset_ComputedUserset:
		rewrittenReq := req.clone()
		rewrittenReq.Relation = rewrite.ComputedUserset.GetRelation()
		resp = l.dispatch(ctx, rewrittenReq, foundUsersChan)
	case *openfgav1.Userset_TupleToUserset:
		resp = l.expandTTU(ctx, req, rewrite, foundUsersChan)
	case *openfgav1.Userset_Intersection:
		resp = l.expandIntersection(ctx, req, rewrite, foundUsersChan)
	case *openfgav1.Userset_Difference:
		resp = l.expandExclusion(ctx, req, rewrite, foundUsersChan)
	case *openfgav1.Userset_Union:
		resp = l.expandUnion(ctx, req, rewrite, foundUsersChan)
	default:
		panic("unexpected userset rewrite encountered")
	}

	if resp.err != nil {
		telemetry.TraceError(span, resp.err)
	}
	return resp
}

func (l *listUsersQuery) expandDirect(
	ctx context.Context,
	req *internalListUsersRequest,
	foundUsersChan chan<- foundUser,
) expandResponse {
	ctx, span := tracer.Start(ctx, "expandDirect")
	defer span.End()
	typesys, err := l.typesystemResolver(ctx, req.GetStoreId(), req.GetAuthorizationModelId())
	if err != nil {
		return expandResponse{
			err: err,
		}
	}

	iter, err := l.ds.Read(ctx, req.GetStoreId(), &openfgav1.TupleKey{
		Object:   tuple.ObjectKey(req.GetObject()),
		Relation: req.GetRelation(),
	})
	if err != nil {
		telemetry.TraceError(span, err)
		return expandResponse{
			err: err,
		}
	}
	defer iter.Stop()
	req.datastoreQueryCount.Add(1)

	filteredIter := storage.NewFilteredTupleKeyIterator(
		storage.NewTupleKeyIteratorFromTupleIterator(iter),
		validation.FilterInvalidTuples(typesys),
	)
	defer filteredIter.Stop()

	pool := pool.New().WithContext(ctx)
	pool.WithCancelOnError()
	pool.WithMaxGoroutines(int(l.resolveNodeBreadthLimit))

	var errs error
	var hasCycle atomic.Bool
LoopOnIterator:
	for {
		tupleKey, err := filteredIter.Next(ctx)
		if err != nil {
			if !errors.Is(err, storage.ErrIteratorDone) {
				errs = errors.Join(errs, err)
			}

			break LoopOnIterator
		}

		condEvalResult, err := eval.EvaluateTupleCondition(ctx, tupleKey, typesys, req.GetContext())
		if err != nil {
			errs = errors.Join(errs, err)
			break LoopOnIterator
		}

		if len(condEvalResult.MissingParameters) > 0 {
			err := condition.NewEvaluationError(
				tupleKey.GetCondition().GetName(),
				fmt.Errorf("context is missing parameters '%v'", condEvalResult.MissingParameters),
			)
			telemetry.TraceError(span, err)
			errs = errors.Join(errs, err)
		}

		if !condEvalResult.ConditionMet {
			continue
		}

		tupleKeyUser := tupleKey.GetUser()
		userObject, userRelation := tuple.SplitObjectRelation(tupleKeyUser)
		userObjectType, userObjectID := tuple.SplitObject(userObject)

		if userRelation == "" {
			for _, f := range req.GetUserFilters() {
				if f.GetType() == userObjectType {
					user := tuple.StringToUserProto(tuple.BuildObject(userObjectType, userObjectID))

					trySendResult(ctx, foundUser{
						user: user,
					}, foundUsersChan)
				}
			}
			continue
		}

		pool.Go(func(ctx context.Context) error {
			rewrittenReq := req.clone()
			rewrittenReq.Object = &openfgav1.Object{Type: userObjectType, Id: userObjectID}
			rewrittenReq.Relation = userRelation
			resp := l.dispatch(ctx, rewrittenReq, foundUsersChan)
			if resp.hasCycle {
				hasCycle.Store(true)
			}
			return resp.err
		})
	}

	errs = errors.Join(errs, pool.Wait())
	if errs != nil {
		telemetry.TraceError(span, errs)
	}
	return expandResponse{
		err:      errs,
		hasCycle: hasCycle.Load(),
	}
}

func (l *listUsersQuery) expandIntersection(
	ctx context.Context,
	req *internalListUsersRequest,
	rewrite *openfgav1.Userset_Intersection,
	foundUsersChan chan<- foundUser,
) expandResponse {
	ctx, span := tracer.Start(ctx, "expandIntersection")
	defer span.End()
	pool := pool.New().WithContext(ctx)
	pool.WithCancelOnError()
	pool.WithMaxGoroutines(int(l.resolveNodeBreadthLimit))

	childOperands := rewrite.Intersection.GetChild()
	intersectionFoundUsersChans := make([]chan foundUser, len(childOperands))
	for i, rewrite := range childOperands {
		i := i
		rewrite := rewrite
		intersectionFoundUsersChans[i] = make(chan foundUser, 1)
		pool.Go(func(ctx context.Context) error {
			resp := l.expandRewrite(ctx, req, rewrite, intersectionFoundUsersChans[i])
			return resp.err
		})
	}

	errChan := make(chan error, 1)

	go func() {
		err := pool.Wait()
		for i := range intersectionFoundUsersChans {
			close(intersectionFoundUsersChans[i])
		}
		errChan <- err
		close(errChan)
	}()

	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(len(childOperands))

	wildcardCount := atomic.Uint32{}
	wildcardKey := tuple.TypedPublicWildcard(req.GetUserFilters()[0].GetType())
	foundUsersCountMap := make(map[string]uint32, 0)
	excludedUsersMap := make(map[string]struct{}, 0)
	for _, foundUsersChan := range intersectionFoundUsersChans {
		go func(foundUsersChan chan foundUser) {
			defer wg.Done()
			foundUsersMap := make(map[string]uint32, 0)
			for foundUser := range foundUsersChan {
				key := tuple.UserProtoToString(foundUser.user)
				for _, excludedUser := range foundUser.excludedUsers {
					key := tuple.UserProtoToString(excludedUser)
					mu.Lock()
					excludedUsersMap[key] = struct{}{}
					mu.Unlock()
				}
				if foundUser.relationshipStatus == NoRelationship {
					continue
				}
				foundUsersMap[key]++
			}

			_, wildcardExists := foundUsersMap[wildcardKey]
			if wildcardExists {
				wildcardCount.Add(1)
			}
			for userKey := range foundUsersMap {
				mu.Lock()
				// Increment the count for a user but decrement if a wildcard
				// also exists to prevent double counting. This ensures accurate
				// tracking for intersection criteria, avoiding inflated counts
				// when both a user and a wildcard are present.
				foundUsersCountMap[userKey]++
				if wildcardExists {
					foundUsersCountMap[userKey]--
				}
				mu.Unlock()
			}
		}(foundUsersChan)
	}
	wg.Wait()

	excludedUsers := []*openfgav1.User{}
	for key := range excludedUsersMap {
		excludedUsers = append(excludedUsers, tuple.StringToUserProto(key))
	}

	for key, count := range foundUsersCountMap {
		// Compare the number of times the specific user was returned for
		// all intersection operands plus the number of wildcards.
		// If this summed value equals the number of operands, the user satisfies
		// the intersection expression and can be sent on `foundUsersChan`
		if (count + wildcardCount.Load()) == uint32(len(childOperands)) {
			fu := foundUser{
				user:          tuple.StringToUserProto(key),
				excludedUsers: excludedUsers,
			}
			trySendResult(ctx, fu, foundUsersChan)
		}
	}

	return expandResponse{
		err: <-errChan,
	}
}

func (l *listUsersQuery) expandUnion(
	ctx context.Context,
	req *internalListUsersRequest,
	rewrite *openfgav1.Userset_Union,
	foundUsersChan chan<- foundUser,
) expandResponse {
	ctx, span := tracer.Start(ctx, "expandUnion")
	defer span.End()
	pool := pool.New().WithContext(ctx)
	pool.WithCancelOnError()
	pool.WithMaxGoroutines(int(l.resolveNodeBreadthLimit))

	childOperands := rewrite.Union.GetChild()
	unionFoundUsersChans := make([]chan foundUser, len(childOperands))
	for i, rewrite := range childOperands {
		i := i
		rewrite := rewrite
		unionFoundUsersChans[i] = make(chan foundUser, 1)
		pool.Go(func(ctx context.Context) error {
			resp := l.expandRewrite(ctx, req, rewrite, unionFoundUsersChans[i])
			return resp.err
		})
	}

	errChan := make(chan error, 1)

	go func() {
		err := pool.Wait()
		for i := range unionFoundUsersChans {
			close(unionFoundUsersChans[i])
		}
		errChan <- err
		close(errChan)
	}()

	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(len(childOperands))

	foundUsersMap := make(map[string]struct{}, 0)
	excludedUsersCountMap := make(map[string]uint32, 0)
	for _, foundUsersChan := range unionFoundUsersChans {
		go func(foundUsersChan chan foundUser) {
			defer wg.Done()

			for foundUser := range foundUsersChan {
				key := tuple.UserProtoToString(foundUser.user)
				for _, excludedUser := range foundUser.excludedUsers {
					key := tuple.UserProtoToString(excludedUser)
					mu.Lock()
					excludedUsersCountMap[key]++
					mu.Unlock()
				}
				if foundUser.relationshipStatus == NoRelationship {
					continue
				}
				mu.Lock()
				foundUsersMap[key] = struct{}{}
				mu.Unlock()
			}
		}(foundUsersChan)
	}
	wg.Wait()

	excludedUsers := []*openfgav1.User{}
	for key, count := range excludedUsersCountMap {
		if count == uint32(len(childOperands)) {
			excludedUsers = append(excludedUsers, tuple.StringToUserProto(key))
		}
	}

	for key := range foundUsersMap {
		fu := foundUser{
			user:          tuple.StringToUserProto(key),
			excludedUsers: excludedUsers,
		}
		trySendResult(ctx, fu, foundUsersChan)
	}

	return expandResponse{
		err: <-errChan,
	}
}

func (l *listUsersQuery) expandExclusion(
	ctx context.Context,
	req *internalListUsersRequest,
	rewrite *openfgav1.Userset_Difference,
	foundUsersChan chan<- foundUser,
) expandResponse {
	ctx, span := tracer.Start(ctx, "expandExclusion")
	defer span.End()
	baseFoundUsersCh := make(chan foundUser, 1)
	subtractFoundUsersCh := make(chan foundUser, 1)

	var baseError error
	go func() {
		resp := l.expandRewrite(ctx, req, rewrite.Difference.GetBase(), baseFoundUsersCh)
		baseError = resp.err
		close(baseFoundUsersCh)
	}()

	var subtractError error
	var subtractHasCycle bool
	go func() {
		resp := l.expandRewrite(ctx, req, rewrite.Difference.GetSubtract(), subtractFoundUsersCh)
		subtractError = resp.err
		subtractHasCycle = resp.hasCycle
		close(subtractFoundUsersCh)
	}()

	baseFoundUsersMap := make(map[string]foundUser, 0)
	for fu := range baseFoundUsersCh {
		key := tuple.UserProtoToString(fu.user)
		baseFoundUsersMap[key] = fu
	}

	subtractFoundUsersMap := make(map[string]foundUser, len(baseFoundUsersMap))

	for fu := range subtractFoundUsersCh {
		key := tuple.UserProtoToString(fu.user)
		subtractFoundUsersMap[key] = fu
	}

	if subtractHasCycle {
		// Because exclusion contains the only bespoke treatment of
		// cycle, everywhere else we consider it a falsey outcome.
		// Once we make a determination within the exclusion handler, we're
		// able to properly handle the case and do not need to propagate
		// the existence of a cycle to an upstream handler.
		return expandResponse{
			err: nil,
		}
	}

	wildcardKey := tuple.TypedPublicWildcard(req.GetUserFilters()[0].GetType())

	_, baseWildcardExists := baseFoundUsersMap[wildcardKey]
	_, subtractWildcardExists := subtractFoundUsersMap[wildcardKey]

	for userKey, fu := range baseFoundUsersMap {
		subtractedUser, userIsSubtracted := subtractFoundUsersMap[userKey]
		_, wildcardSubtracted := subtractFoundUsersMap[wildcardKey]

		switch {
		case baseWildcardExists:
			if !userIsSubtracted && !wildcardSubtracted {
				trySendResult(ctx, foundUser{
					user: tuple.StringToUserProto(userKey),
				}, foundUsersChan)
			}

			for subtractedUserKey, subtractedFu := range subtractFoundUsersMap {
				if tuple.IsTypedWildcard(subtractedUserKey) {
					if !userIsSubtracted {
						trySendResult(ctx, foundUser{
							user:               tuple.StringToUserProto(userKey),
							relationshipStatus: NoRelationship,
						}, foundUsersChan)
					}
					continue
				}

				if subtractedFu.relationshipStatus == NoRelationship {
					trySendResult(ctx, foundUser{
						user:               tuple.StringToUserProto(subtractedUserKey),
						relationshipStatus: HasRelationship,
					}, foundUsersChan)
				}

				// a found user under the subtracted branch causes the subtracted user to have a negated relationship with respect
				// to the base relation and is excluded since a wildcard is contained under the base branch.
				if subtractedFu.relationshipStatus == HasRelationship {
					trySendResult(ctx, foundUser{
						user:               tuple.StringToUserProto(subtractedUserKey),
						relationshipStatus: NoRelationship,
						excludedUsers: []*openfgav1.User{
							tuple.StringToUserProto(subtractedUserKey),
						},
					}, foundUsersChan)
				}
			}
		case subtractWildcardExists, userIsSubtracted:
			if subtractedUser.relationshipStatus == HasRelationship {
				trySendResult(ctx, foundUser{
					user:               tuple.StringToUserProto(userKey),
					relationshipStatus: NoRelationship,
				}, foundUsersChan)
			}

			if subtractedUser.relationshipStatus == NoRelationship {
				trySendResult(ctx, foundUser{
					user:               tuple.StringToUserProto(userKey),
					relationshipStatus: HasRelationship,
				}, foundUsersChan)
			}

		default:
			trySendResult(ctx, foundUser{
				user:               tuple.StringToUserProto(userKey),
				relationshipStatus: fu.relationshipStatus,
			}, foundUsersChan)
		}
	}

	errs := errors.Join(baseError, subtractError)
	if errs != nil {
		telemetry.TraceError(span, errs)
	}
	return expandResponse{
		err: errs,
	}
}

func (l *listUsersQuery) expandTTU(
	ctx context.Context,
	req *internalListUsersRequest,
	rewrite *openfgav1.Userset_TupleToUserset,
	foundUsersChan chan<- foundUser,
) expandResponse {
	ctx, span := tracer.Start(ctx, "expandTTU")
	defer span.End()
	tuplesetRelation := rewrite.TupleToUserset.GetTupleset().GetRelation()
	computedRelation := rewrite.TupleToUserset.GetComputedUserset().GetRelation()

	typesys, err := l.typesystemResolver(ctx, req.GetStoreId(), req.GetAuthorizationModelId())
	if err != nil {
		return expandResponse{
			err: err,
		}
	}

	iter, err := l.ds.Read(ctx, req.GetStoreId(), &openfgav1.TupleKey{
		Object:   tuple.ObjectKey(req.GetObject()),
		Relation: tuplesetRelation,
	})
	if err != nil {
		telemetry.TraceError(span, err)
		return expandResponse{
			err: err,
		}
	}
	defer iter.Stop()
	req.datastoreQueryCount.Add(1)

	filteredIter := storage.NewFilteredTupleKeyIterator(
		storage.NewTupleKeyIteratorFromTupleIterator(iter),
		validation.FilterInvalidTuples(typesys),
	)
	defer filteredIter.Stop()

	pool := pool.New().WithContext(ctx)
	pool.WithCancelOnError()
	pool.WithMaxGoroutines(int(l.resolveNodeBreadthLimit))

	var errs error

LoopOnIterator:
	for {
		tupleKey, err := filteredIter.Next(ctx)
		if err != nil {
			if !errors.Is(err, storage.ErrIteratorDone) {
				errs = errors.Join(errs, err)
			}

			break LoopOnIterator
		}

		condEvalResult, err := eval.EvaluateTupleCondition(ctx, tupleKey, typesys, req.GetContext())
		if err != nil {
			errs = errors.Join(errs, err)
			break LoopOnIterator
		}

		if len(condEvalResult.MissingParameters) > 0 {
			err := condition.NewEvaluationError(
				tupleKey.GetCondition().GetName(),
				fmt.Errorf("context is missing parameters '%v'", condEvalResult.MissingParameters),
			)
			telemetry.TraceError(span, err)
			errs = errors.Join(errs, err)
		}

		if !condEvalResult.ConditionMet {
			continue
		}

		userObject := tupleKey.GetUser()
		userObjectType, userObjectID := tuple.SplitObject(userObject)

		pool.Go(func(ctx context.Context) error {
			rewrittenReq := req.clone()
			rewrittenReq.Object = &openfgav1.Object{Type: userObjectType, Id: userObjectID}
			rewrittenReq.Relation = computedRelation
			resp := l.dispatch(ctx, rewrittenReq, foundUsersChan)
			return resp.err
		})
	}

	errs = errors.Join(pool.Wait(), errs)
	if errs != nil {
		telemetry.TraceError(span, errs)
	}
	return expandResponse{
		err: errs,
	}
}

func enteredCycle(req *internalListUsersRequest) bool {
	key := fmt.Sprintf("%s#%s", tuple.ObjectKey(req.GetObject()), req.Relation)
	if _, loaded := req.visitedUsersetsMap[key]; loaded {
		return true
	}
	req.visitedUsersetsMap[key] = struct{}{}
	return false
}

func (l *listUsersQuery) buildResultsChannel() chan foundUser {
	foundUsersCh := make(chan foundUser, serverconfig.DefaultListUsersMaxResults)
	maxResults := l.maxResults
	if maxResults > 0 {
		foundUsersCh = make(chan foundUser, maxResults)
	}

	return foundUsersCh
}

func trySendResult(ctx context.Context, user foundUser, foundUsersCh chan<- foundUser) {
	select {
	case <-ctx.Done():
		return
	case foundUsersCh <- user:
		return
	}
}
