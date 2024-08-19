package listusers

import (
	"maps"
	"sync/atomic"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type listUsersRequest interface {
	GetStoreId() string
	GetAuthorizationModelId() string
	GetObject() *openfgav1.Object
	GetRelation() string
	GetUserFilters() []*openfgav1.UserTypeFilter
	GetContextualTuples() []*openfgav1.TupleKey
	GetContext() *structpb.Struct
}

type internalListUsersRequest struct {
	*openfgav1.ListUsersRequest

	// visitedUsersetsMap keeps track of the "path" we've made so far.
	// It prevents stack overflows by preventing visiting the same userset twice.
	visitedUsersetsMap map[string]struct{}

	// depth is the current depths of the traversal expressed as a positive, incrementing integer.
	// When expansion of list users recursively traverses one level, we increment by one. If this
	// counter hits the limit, we throw ErrResolutionDepthExceeded. This protects against a potentially deep
	// or endless cycle of recursion.
	depth uint32

	datastoreQueryCount *atomic.Uint32

	dispatchCount *atomic.Uint32
}

var _ listUsersRequest = (*internalListUsersRequest)(nil)

// nolint // it should be GetStoreID, but we want to satisfy the interface listUsersRequest
func (r *internalListUsersRequest) GetStoreId() string {
	if r == nil {
		return ""
	}
	return r.StoreId
}

// nolint // it should be GetAuthorizationModelID, but we want to satisfy the interface listUsersRequest
func (r *internalListUsersRequest) GetAuthorizationModelId() string {
	if r == nil {
		return ""
	}
	return r.AuthorizationModelId
}

func (r *internalListUsersRequest) GetObject() *openfgav1.Object {
	if r == nil {
		return nil
	}
	return r.Object
}

func (r *internalListUsersRequest) GetRelation() string {
	if r == nil {
		return ""
	}
	return r.Relation
}

func (r *internalListUsersRequest) GetUserFilters() []*openfgav1.UserTypeFilter {
	if r == nil {
		return nil
	}
	return r.UserFilters
}

func (r *internalListUsersRequest) GetContextualTuples() []*openfgav1.TupleKey {
	if r == nil {
		return nil
	}
	return r.ContextualTuples
}

func (r *internalListUsersRequest) GetDatastoreQueryCount() uint32 {
	if r == nil {
		return uint32(0)
	}
	return r.datastoreQueryCount.Load()
}

func (r *internalListUsersRequest) GetDispatchCount() uint32 {
	if r == nil {
		return uint32(0)
	}
	return r.dispatchCount.Load()
}

func (r *internalListUsersRequest) GetContext() *structpb.Struct {
	if r == nil {
		return nil
	}
	return r.Context
}

type listUsersResponse struct {
	Users         []*openfgav1.User
	ExcludedUsers []*openfgav1.ObjectOrUserset
	Metadata      listUsersResponseMetadata
}

type listUsersResponseMetadata struct {
	DatastoreQueryCount uint32

	// The number of times we are recursively expanding to find users.
	// Atomic is used to be consistent with the Check and ListObjects.
	DispatchCounter *atomic.Uint32
}

func (r *listUsersResponse) GetUsers() []*openfgav1.User {
	if r == nil {
		return []*openfgav1.User{}
	}
	return r.Users
}

func (r *listUsersResponse) GetExcludedUsers() []*openfgav1.ObjectOrUserset {
	if r == nil {
		return []*openfgav1.ObjectOrUserset{}
	}
	return r.ExcludedUsers
}

func (r *listUsersResponse) GetMetadata() listUsersResponseMetadata {
	if r == nil {
		return listUsersResponseMetadata{}
	}
	return r.Metadata
}

func fromListUsersRequest(o listUsersRequest, datastoreQueryCount *atomic.Uint32, dispatchCount *atomic.Uint32) *internalListUsersRequest {
	if datastoreQueryCount == nil {
		datastoreQueryCount = new(atomic.Uint32)
	}
	if dispatchCount == nil {
		dispatchCount = new(atomic.Uint32)
	}
	return &internalListUsersRequest{
		ListUsersRequest: &openfgav1.ListUsersRequest{
			StoreId:              o.GetStoreId(),
			AuthorizationModelId: o.GetAuthorizationModelId(),
			Object:               o.GetObject(),
			Relation:             o.GetRelation(),
			UserFilters:          o.GetUserFilters(),
			ContextualTuples:     o.GetContextualTuples(),
			Context:              o.GetContext(),
		},
		visitedUsersetsMap:  make(map[string]struct{}),
		depth:               0,
		datastoreQueryCount: datastoreQueryCount,
		dispatchCount:       dispatchCount,
	}
}

// clone creates a copy of the request. Note that some fields are not deep-cloned.
func (r *internalListUsersRequest) clone() *internalListUsersRequest {
	v := fromListUsersRequest(r, r.datastoreQueryCount, r.dispatchCount)
	v.visitedUsersetsMap = maps.Clone(r.visitedUsersetsMap)
	v.depth = r.depth
	return v
}
