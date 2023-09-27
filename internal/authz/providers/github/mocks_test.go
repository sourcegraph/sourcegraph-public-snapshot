// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge github

import (
	"context"
	"sync"

	buth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	github "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
)

// MockClient is b mock implementbtion of the client interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/github) used
// for unit testing.
type MockClient struct {
	// GetAuthenticbtedOAuthScopesFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetAuthenticbtedOAuthScopes.
	GetAuthenticbtedOAuthScopesFunc *ClientGetAuthenticbtedOAuthScopesFunc
	// GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc is bn instbnce of b
	// mock function object controlling the behbvior of the method
	// GetAuthenticbtedUserOrgsDetbilsAndMembership.
	GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc *ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc
	// GetAuthenticbtedUserTebmsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetAuthenticbtedUserTebms.
	GetAuthenticbtedUserTebmsFunc *ClientGetAuthenticbtedUserTebmsFunc
	// GetOrgbnizbtionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetOrgbnizbtion.
	GetOrgbnizbtionFunc *ClientGetOrgbnizbtionFunc
	// GetRepositoryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRepository.
	GetRepositoryFunc *ClientGetRepositoryFunc
	// ListAffilibtedRepositoriesFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// ListAffilibtedRepositories.
	ListAffilibtedRepositoriesFunc *ClientListAffilibtedRepositoriesFunc
	// ListOrgRepositoriesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListOrgRepositories.
	ListOrgRepositoriesFunc *ClientListOrgRepositoriesFunc
	// ListOrgbnizbtionMembersFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListOrgbnizbtionMembers.
	ListOrgbnizbtionMembersFunc *ClientListOrgbnizbtionMembersFunc
	// ListRepositoryCollbborbtorsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// ListRepositoryCollbborbtors.
	ListRepositoryCollbborbtorsFunc *ClientListRepositoryCollbborbtorsFunc
	// ListRepositoryTebmsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListRepositoryTebms.
	ListRepositoryTebmsFunc *ClientListRepositoryTebmsFunc
	// ListTebmMembersFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListTebmMembers.
	ListTebmMembersFunc *ClientListTebmMembersFunc
	// ListTebmRepositoriesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListTebmRepositories.
	ListTebmRepositoriesFunc *ClientListTebmRepositoriesFunc
	// SetWbitForRbteLimitFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetWbitForRbteLimit.
	SetWbitForRbteLimitFunc *ClientSetWbitForRbteLimitFunc
	// WithAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithAuthenticbtor.
	WithAuthenticbtorFunc *ClientWithAuthenticbtorFunc
}

// NewMockClient crebtes b new mock of the client interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockClient() *MockClient {
	return &MockClient{
		GetAuthenticbtedOAuthScopesFunc: &ClientGetAuthenticbtedOAuthScopesFunc{
			defbultHook: func(context.Context) (r0 []string, r1 error) {
				return
			},
		},
		GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc: &ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc{
			defbultHook: func(context.Context, int) (r0 []github.OrgDetbilsAndMembership, r1 bool, r2 int, r3 error) {
				return
			},
		},
		GetAuthenticbtedUserTebmsFunc: &ClientGetAuthenticbtedUserTebmsFunc{
			defbultHook: func(context.Context, int) (r0 []*github.Tebm, r1 bool, r2 int, r3 error) {
				return
			},
		},
		GetOrgbnizbtionFunc: &ClientGetOrgbnizbtionFunc{
			defbultHook: func(context.Context, string) (r0 *github.OrgDetbils, r1 error) {
				return
			},
		},
		GetRepositoryFunc: &ClientGetRepositoryFunc{
			defbultHook: func(context.Context, string, string) (r0 *github.Repository, r1 error) {
				return
			},
		},
		ListAffilibtedRepositoriesFunc: &ClientListAffilibtedRepositoriesFunc{
			defbultHook: func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) (r0 []*github.Repository, r1 bool, r2 int, r3 error) {
				return
			},
		},
		ListOrgRepositoriesFunc: &ClientListOrgRepositoriesFunc{
			defbultHook: func(context.Context, string, int, string) (r0 []*github.Repository, r1 bool, r2 int, r3 error) {
				return
			},
		},
		ListOrgbnizbtionMembersFunc: &ClientListOrgbnizbtionMembersFunc{
			defbultHook: func(context.Context, string, int, bool) (r0 []*github.Collbborbtor, r1 bool, r2 error) {
				return
			},
		},
		ListRepositoryCollbborbtorsFunc: &ClientListRepositoryCollbborbtorsFunc{
			defbultHook: func(context.Context, string, string, int, github.CollbborbtorAffilibtion) (r0 []*github.Collbborbtor, r1 bool, r2 error) {
				return
			},
		},
		ListRepositoryTebmsFunc: &ClientListRepositoryTebmsFunc{
			defbultHook: func(context.Context, string, string, int) (r0 []*github.Tebm, r1 bool, r2 error) {
				return
			},
		},
		ListTebmMembersFunc: &ClientListTebmMembersFunc{
			defbultHook: func(context.Context, string, string, int) (r0 []*github.Collbborbtor, r1 bool, r2 error) {
				return
			},
		},
		ListTebmRepositoriesFunc: &ClientListTebmRepositoriesFunc{
			defbultHook: func(context.Context, string, string, int) (r0 []*github.Repository, r1 bool, r2 int, r3 error) {
				return
			},
		},
		SetWbitForRbteLimitFunc: &ClientSetWbitForRbteLimitFunc{
			defbultHook: func(bool) {
				return
			},
		},
		WithAuthenticbtorFunc: &ClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (r0 client) {
				return
			},
		},
	}
}

// NewStrictMockClient crebtes b new mock of the client interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockClient() *MockClient {
	return &MockClient{
		GetAuthenticbtedOAuthScopesFunc: &ClientGetAuthenticbtedOAuthScopesFunc{
			defbultHook: func(context.Context) ([]string, error) {
				pbnic("unexpected invocbtion of MockClient.GetAuthenticbtedOAuthScopes")
			},
		},
		GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc: &ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc{
			defbultHook: func(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error) {
				pbnic("unexpected invocbtion of MockClient.GetAuthenticbtedUserOrgsDetbilsAndMembership")
			},
		},
		GetAuthenticbtedUserTebmsFunc: &ClientGetAuthenticbtedUserTebmsFunc{
			defbultHook: func(context.Context, int) ([]*github.Tebm, bool, int, error) {
				pbnic("unexpected invocbtion of MockClient.GetAuthenticbtedUserTebms")
			},
		},
		GetOrgbnizbtionFunc: &ClientGetOrgbnizbtionFunc{
			defbultHook: func(context.Context, string) (*github.OrgDetbils, error) {
				pbnic("unexpected invocbtion of MockClient.GetOrgbnizbtion")
			},
		},
		GetRepositoryFunc: &ClientGetRepositoryFunc{
			defbultHook: func(context.Context, string, string) (*github.Repository, error) {
				pbnic("unexpected invocbtion of MockClient.GetRepository")
			},
		},
		ListAffilibtedRepositoriesFunc: &ClientListAffilibtedRepositoriesFunc{
			defbultHook: func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error) {
				pbnic("unexpected invocbtion of MockClient.ListAffilibtedRepositories")
			},
		},
		ListOrgRepositoriesFunc: &ClientListOrgRepositoriesFunc{
			defbultHook: func(context.Context, string, int, string) ([]*github.Repository, bool, int, error) {
				pbnic("unexpected invocbtion of MockClient.ListOrgRepositories")
			},
		},
		ListOrgbnizbtionMembersFunc: &ClientListOrgbnizbtionMembersFunc{
			defbultHook: func(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error) {
				pbnic("unexpected invocbtion of MockClient.ListOrgbnizbtionMembers")
			},
		},
		ListRepositoryCollbborbtorsFunc: &ClientListRepositoryCollbborbtorsFunc{
			defbultHook: func(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error) {
				pbnic("unexpected invocbtion of MockClient.ListRepositoryCollbborbtors")
			},
		},
		ListRepositoryTebmsFunc: &ClientListRepositoryTebmsFunc{
			defbultHook: func(context.Context, string, string, int) ([]*github.Tebm, bool, error) {
				pbnic("unexpected invocbtion of MockClient.ListRepositoryTebms")
			},
		},
		ListTebmMembersFunc: &ClientListTebmMembersFunc{
			defbultHook: func(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error) {
				pbnic("unexpected invocbtion of MockClient.ListTebmMembers")
			},
		},
		ListTebmRepositoriesFunc: &ClientListTebmRepositoriesFunc{
			defbultHook: func(context.Context, string, string, int) ([]*github.Repository, bool, int, error) {
				pbnic("unexpected invocbtion of MockClient.ListTebmRepositories")
			},
		},
		SetWbitForRbteLimitFunc: &ClientSetWbitForRbteLimitFunc{
			defbultHook: func(bool) {
				pbnic("unexpected invocbtion of MockClient.SetWbitForRbteLimit")
			},
		},
		WithAuthenticbtorFunc: &ClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) client {
				pbnic("unexpected invocbtion of MockClient.WithAuthenticbtor")
			},
		},
	}
}

// surrogbteMockClient is b copy of the client interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/github). It
// is redefined here bs it is unexported in the source pbckbge.
type surrogbteMockClient interfbce {
	GetAuthenticbtedOAuthScopes(context.Context) ([]string, error)
	GetAuthenticbtedUserOrgsDetbilsAndMembership(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error)
	GetAuthenticbtedUserTebms(context.Context, int) ([]*github.Tebm, bool, int, error)
	GetOrgbnizbtion(context.Context, string) (*github.OrgDetbils, error)
	GetRepository(context.Context, string, string) (*github.Repository, error)
	ListAffilibtedRepositories(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error)
	ListOrgRepositories(context.Context, string, int, string) ([]*github.Repository, bool, int, error)
	ListOrgbnizbtionMembers(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error)
	ListRepositoryCollbborbtors(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error)
	ListRepositoryTebms(context.Context, string, string, int) ([]*github.Tebm, bool, error)
	ListTebmMembers(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error)
	ListTebmRepositories(context.Context, string, string, int) ([]*github.Repository, bool, int, error)
	SetWbitForRbteLimit(bool)
	WithAuthenticbtor(buth.Authenticbtor) client
}

// NewMockClientFrom crebtes b new mock of the MockClient interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockClientFrom(i surrogbteMockClient) *MockClient {
	return &MockClient{
		GetAuthenticbtedOAuthScopesFunc: &ClientGetAuthenticbtedOAuthScopesFunc{
			defbultHook: i.GetAuthenticbtedOAuthScopes,
		},
		GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc: &ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc{
			defbultHook: i.GetAuthenticbtedUserOrgsDetbilsAndMembership,
		},
		GetAuthenticbtedUserTebmsFunc: &ClientGetAuthenticbtedUserTebmsFunc{
			defbultHook: i.GetAuthenticbtedUserTebms,
		},
		GetOrgbnizbtionFunc: &ClientGetOrgbnizbtionFunc{
			defbultHook: i.GetOrgbnizbtion,
		},
		GetRepositoryFunc: &ClientGetRepositoryFunc{
			defbultHook: i.GetRepository,
		},
		ListAffilibtedRepositoriesFunc: &ClientListAffilibtedRepositoriesFunc{
			defbultHook: i.ListAffilibtedRepositories,
		},
		ListOrgRepositoriesFunc: &ClientListOrgRepositoriesFunc{
			defbultHook: i.ListOrgRepositories,
		},
		ListOrgbnizbtionMembersFunc: &ClientListOrgbnizbtionMembersFunc{
			defbultHook: i.ListOrgbnizbtionMembers,
		},
		ListRepositoryCollbborbtorsFunc: &ClientListRepositoryCollbborbtorsFunc{
			defbultHook: i.ListRepositoryCollbborbtors,
		},
		ListRepositoryTebmsFunc: &ClientListRepositoryTebmsFunc{
			defbultHook: i.ListRepositoryTebms,
		},
		ListTebmMembersFunc: &ClientListTebmMembersFunc{
			defbultHook: i.ListTebmMembers,
		},
		ListTebmRepositoriesFunc: &ClientListTebmRepositoriesFunc{
			defbultHook: i.ListTebmRepositories,
		},
		SetWbitForRbteLimitFunc: &ClientSetWbitForRbteLimitFunc{
			defbultHook: i.SetWbitForRbteLimit,
		},
		WithAuthenticbtorFunc: &ClientWithAuthenticbtorFunc{
			defbultHook: i.WithAuthenticbtor,
		},
	}
}

// ClientGetAuthenticbtedOAuthScopesFunc describes the behbvior when the
// GetAuthenticbtedOAuthScopes method of the pbrent MockClient instbnce is
// invoked.
type ClientGetAuthenticbtedOAuthScopesFunc struct {
	defbultHook func(context.Context) ([]string, error)
	hooks       []func(context.Context) ([]string, error)
	history     []ClientGetAuthenticbtedOAuthScopesFuncCbll
	mutex       sync.Mutex
}

// GetAuthenticbtedOAuthScopes delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetAuthenticbtedOAuthScopes(v0 context.Context) ([]string, error) {
	r0, r1 := m.GetAuthenticbtedOAuthScopesFunc.nextHook()(v0)
	m.GetAuthenticbtedOAuthScopesFunc.bppendCbll(ClientGetAuthenticbtedOAuthScopesFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetAuthenticbtedOAuthScopes method of the pbrent MockClient instbnce is
// invoked bnd the hook queue is empty.
func (f *ClientGetAuthenticbtedOAuthScopesFunc) SetDefbultHook(hook func(context.Context) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAuthenticbtedOAuthScopes method of the pbrent MockClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ClientGetAuthenticbtedOAuthScopesFunc) PushHook(hook func(context.Context) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetAuthenticbtedOAuthScopesFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetAuthenticbtedOAuthScopesFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context) ([]string, error) {
		return r0, r1
	})
}

func (f *ClientGetAuthenticbtedOAuthScopesFunc) nextHook() func(context.Context) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetAuthenticbtedOAuthScopesFunc) bppendCbll(r0 ClientGetAuthenticbtedOAuthScopesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetAuthenticbtedOAuthScopesFuncCbll
// objects describing the invocbtions of this function.
func (f *ClientGetAuthenticbtedOAuthScopesFunc) History() []ClientGetAuthenticbtedOAuthScopesFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetAuthenticbtedOAuthScopesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetAuthenticbtedOAuthScopesFuncCbll is bn object thbt describes bn
// invocbtion of method GetAuthenticbtedOAuthScopes on bn instbnce of
// MockClient.
type ClientGetAuthenticbtedOAuthScopesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetAuthenticbtedOAuthScopesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetAuthenticbtedOAuthScopesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc describes the
// behbvior when the GetAuthenticbtedUserOrgsDetbilsAndMembership method of
// the pbrent MockClient instbnce is invoked.
type ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc struct {
	defbultHook func(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error)
	hooks       []func(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error)
	history     []ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll
	mutex       sync.Mutex
}

// GetAuthenticbtedUserOrgsDetbilsAndMembership delegbtes to the next hook
// function in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockClient) GetAuthenticbtedUserOrgsDetbilsAndMembership(v0 context.Context, v1 int) ([]github.OrgDetbilsAndMembership, bool, int, error) {
	r0, r1, r2, r3 := m.GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc.nextHook()(v0, v1)
	m.GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc.bppendCbll(ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll{v0, v1, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the
// GetAuthenticbtedUserOrgsDetbilsAndMembership method of the pbrent
// MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc) SetDefbultHook(hook func(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAuthenticbtedUserOrgsDetbilsAndMembership method of the pbrent
// MockClient instbnce invokes the hook bt the front of the queue bnd
// discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc) PushHook(hook func(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc) SetDefbultReturn(r0 []github.OrgDetbilsAndMembership, r1 bool, r2 int, r3 error) {
	f.SetDefbultHook(func(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc) PushReturn(r0 []github.OrgDetbilsAndMembership, r1 bool, r2 int, r3 error) {
	f.PushHook(func(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error) {
		return r0, r1, r2, r3
	})
}

func (f *ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc) nextHook() func(context.Context, int) ([]github.OrgDetbilsAndMembership, bool, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc) bppendCbll(r0 ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFunc) History() []ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll is bn object
// thbt describes bn invocbtion of method
// GetAuthenticbtedUserOrgsDetbilsAndMembership on bn instbnce of
// MockClient.
type ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []github.OrgDetbilsAndMembership
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 int
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetAuthenticbtedUserOrgsDetbilsAndMembershipFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// ClientGetAuthenticbtedUserTebmsFunc describes the behbvior when the
// GetAuthenticbtedUserTebms method of the pbrent MockClient instbnce is
// invoked.
type ClientGetAuthenticbtedUserTebmsFunc struct {
	defbultHook func(context.Context, int) ([]*github.Tebm, bool, int, error)
	hooks       []func(context.Context, int) ([]*github.Tebm, bool, int, error)
	history     []ClientGetAuthenticbtedUserTebmsFuncCbll
	mutex       sync.Mutex
}

// GetAuthenticbtedUserTebms delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetAuthenticbtedUserTebms(v0 context.Context, v1 int) ([]*github.Tebm, bool, int, error) {
	r0, r1, r2, r3 := m.GetAuthenticbtedUserTebmsFunc.nextHook()(v0, v1)
	m.GetAuthenticbtedUserTebmsFunc.bppendCbll(ClientGetAuthenticbtedUserTebmsFuncCbll{v0, v1, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the
// GetAuthenticbtedUserTebms method of the pbrent MockClient instbnce is
// invoked bnd the hook queue is empty.
func (f *ClientGetAuthenticbtedUserTebmsFunc) SetDefbultHook(hook func(context.Context, int) ([]*github.Tebm, bool, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAuthenticbtedUserTebms method of the pbrent MockClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ClientGetAuthenticbtedUserTebmsFunc) PushHook(hook func(context.Context, int) ([]*github.Tebm, bool, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetAuthenticbtedUserTebmsFunc) SetDefbultReturn(r0 []*github.Tebm, r1 bool, r2 int, r3 error) {
	f.SetDefbultHook(func(context.Context, int) ([]*github.Tebm, bool, int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetAuthenticbtedUserTebmsFunc) PushReturn(r0 []*github.Tebm, r1 bool, r2 int, r3 error) {
	f.PushHook(func(context.Context, int) ([]*github.Tebm, bool, int, error) {
		return r0, r1, r2, r3
	})
}

func (f *ClientGetAuthenticbtedUserTebmsFunc) nextHook() func(context.Context, int) ([]*github.Tebm, bool, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetAuthenticbtedUserTebmsFunc) bppendCbll(r0 ClientGetAuthenticbtedUserTebmsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetAuthenticbtedUserTebmsFuncCbll
// objects describing the invocbtions of this function.
func (f *ClientGetAuthenticbtedUserTebmsFunc) History() []ClientGetAuthenticbtedUserTebmsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetAuthenticbtedUserTebmsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetAuthenticbtedUserTebmsFuncCbll is bn object thbt describes bn
// invocbtion of method GetAuthenticbtedUserTebms on bn instbnce of
// MockClient.
type ClientGetAuthenticbtedUserTebmsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Tebm
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 int
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetAuthenticbtedUserTebmsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetAuthenticbtedUserTebmsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// ClientGetOrgbnizbtionFunc describes the behbvior when the GetOrgbnizbtion
// method of the pbrent MockClient instbnce is invoked.
type ClientGetOrgbnizbtionFunc struct {
	defbultHook func(context.Context, string) (*github.OrgDetbils, error)
	hooks       []func(context.Context, string) (*github.OrgDetbils, error)
	history     []ClientGetOrgbnizbtionFuncCbll
	mutex       sync.Mutex
}

// GetOrgbnizbtion delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetOrgbnizbtion(v0 context.Context, v1 string) (*github.OrgDetbils, error) {
	r0, r1 := m.GetOrgbnizbtionFunc.nextHook()(v0, v1)
	m.GetOrgbnizbtionFunc.bppendCbll(ClientGetOrgbnizbtionFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetOrgbnizbtion
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientGetOrgbnizbtionFunc) SetDefbultHook(hook func(context.Context, string) (*github.OrgDetbils, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetOrgbnizbtion method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientGetOrgbnizbtionFunc) PushHook(hook func(context.Context, string) (*github.OrgDetbils, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetOrgbnizbtionFunc) SetDefbultReturn(r0 *github.OrgDetbils, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*github.OrgDetbils, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetOrgbnizbtionFunc) PushReturn(r0 *github.OrgDetbils, r1 error) {
	f.PushHook(func(context.Context, string) (*github.OrgDetbils, error) {
		return r0, r1
	})
}

func (f *ClientGetOrgbnizbtionFunc) nextHook() func(context.Context, string) (*github.OrgDetbils, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetOrgbnizbtionFunc) bppendCbll(r0 ClientGetOrgbnizbtionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetOrgbnizbtionFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientGetOrgbnizbtionFunc) History() []ClientGetOrgbnizbtionFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetOrgbnizbtionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetOrgbnizbtionFuncCbll is bn object thbt describes bn invocbtion
// of method GetOrgbnizbtion on bn instbnce of MockClient.
type ClientGetOrgbnizbtionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *github.OrgDetbils
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetOrgbnizbtionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetOrgbnizbtionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientGetRepositoryFunc describes the behbvior when the GetRepository
// method of the pbrent MockClient instbnce is invoked.
type ClientGetRepositoryFunc struct {
	defbultHook func(context.Context, string, string) (*github.Repository, error)
	hooks       []func(context.Context, string, string) (*github.Repository, error)
	history     []ClientGetRepositoryFuncCbll
	mutex       sync.Mutex
}

// GetRepository delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetRepository(v0 context.Context, v1 string, v2 string) (*github.Repository, error) {
	r0, r1 := m.GetRepositoryFunc.nextHook()(v0, v1, v2)
	m.GetRepositoryFunc.bppendCbll(ClientGetRepositoryFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetRepository method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientGetRepositoryFunc) SetDefbultHook(hook func(context.Context, string, string) (*github.Repository, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRepository method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientGetRepositoryFunc) PushHook(hook func(context.Context, string, string) (*github.Repository, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetRepositoryFunc) SetDefbultReturn(r0 *github.Repository, r1 error) {
	f.SetDefbultHook(func(context.Context, string, string) (*github.Repository, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetRepositoryFunc) PushReturn(r0 *github.Repository, r1 error) {
	f.PushHook(func(context.Context, string, string) (*github.Repository, error) {
		return r0, r1
	})
}

func (f *ClientGetRepositoryFunc) nextHook() func(context.Context, string, string) (*github.Repository, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetRepositoryFunc) bppendCbll(r0 ClientGetRepositoryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetRepositoryFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientGetRepositoryFunc) History() []ClientGetRepositoryFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetRepositoryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetRepositoryFuncCbll is bn object thbt describes bn invocbtion of
// method GetRepository on bn instbnce of MockClient.
type ClientGetRepositoryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *github.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetRepositoryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetRepositoryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientListAffilibtedRepositoriesFunc describes the behbvior when the
// ListAffilibtedRepositories method of the pbrent MockClient instbnce is
// invoked.
type ClientListAffilibtedRepositoriesFunc struct {
	defbultHook func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error)
	hooks       []func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error)
	history     []ClientListAffilibtedRepositoriesFuncCbll
	mutex       sync.Mutex
}

// ListAffilibtedRepositories delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListAffilibtedRepositories(v0 context.Context, v1 github.Visibility, v2 int, v3 int, v4 ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error) {
	r0, r1, r2, r3 := m.ListAffilibtedRepositoriesFunc.nextHook()(v0, v1, v2, v3, v4...)
	m.ListAffilibtedRepositoriesFunc.bppendCbll(ClientListAffilibtedRepositoriesFuncCbll{v0, v1, v2, v3, v4, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the
// ListAffilibtedRepositories method of the pbrent MockClient instbnce is
// invoked bnd the hook queue is empty.
func (f *ClientListAffilibtedRepositoriesFunc) SetDefbultHook(hook func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListAffilibtedRepositories method of the pbrent MockClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ClientListAffilibtedRepositoriesFunc) PushHook(hook func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListAffilibtedRepositoriesFunc) SetDefbultReturn(r0 []*github.Repository, r1 bool, r2 int, r3 error) {
	f.SetDefbultHook(func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListAffilibtedRepositoriesFunc) PushReturn(r0 []*github.Repository, r1 bool, r2 int, r3 error) {
	f.PushHook(func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error) {
		return r0, r1, r2, r3
	})
}

func (f *ClientListAffilibtedRepositoriesFunc) nextHook() func(context.Context, github.Visibility, int, int, ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListAffilibtedRepositoriesFunc) bppendCbll(r0 ClientListAffilibtedRepositoriesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListAffilibtedRepositoriesFuncCbll
// objects describing the invocbtions of this function.
func (f *ClientListAffilibtedRepositoriesFunc) History() []ClientListAffilibtedRepositoriesFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListAffilibtedRepositoriesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListAffilibtedRepositoriesFuncCbll is bn object thbt describes bn
// invocbtion of method ListAffilibtedRepositories on bn instbnce of
// MockClient.
type ClientListAffilibtedRepositoriesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 github.Visibility
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Arg4 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg4 []github.RepositoryAffilibtion
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 int
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c ClientListAffilibtedRepositoriesFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg4 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListAffilibtedRepositoriesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// ClientListOrgRepositoriesFunc describes the behbvior when the
// ListOrgRepositories method of the pbrent MockClient instbnce is invoked.
type ClientListOrgRepositoriesFunc struct {
	defbultHook func(context.Context, string, int, string) ([]*github.Repository, bool, int, error)
	hooks       []func(context.Context, string, int, string) ([]*github.Repository, bool, int, error)
	history     []ClientListOrgRepositoriesFuncCbll
	mutex       sync.Mutex
}

// ListOrgRepositories delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListOrgRepositories(v0 context.Context, v1 string, v2 int, v3 string) ([]*github.Repository, bool, int, error) {
	r0, r1, r2, r3 := m.ListOrgRepositoriesFunc.nextHook()(v0, v1, v2, v3)
	m.ListOrgRepositoriesFunc.bppendCbll(ClientListOrgRepositoriesFuncCbll{v0, v1, v2, v3, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the ListOrgRepositories
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientListOrgRepositoriesFunc) SetDefbultHook(hook func(context.Context, string, int, string) ([]*github.Repository, bool, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListOrgRepositories method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientListOrgRepositoriesFunc) PushHook(hook func(context.Context, string, int, string) ([]*github.Repository, bool, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListOrgRepositoriesFunc) SetDefbultReturn(r0 []*github.Repository, r1 bool, r2 int, r3 error) {
	f.SetDefbultHook(func(context.Context, string, int, string) ([]*github.Repository, bool, int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListOrgRepositoriesFunc) PushReturn(r0 []*github.Repository, r1 bool, r2 int, r3 error) {
	f.PushHook(func(context.Context, string, int, string) ([]*github.Repository, bool, int, error) {
		return r0, r1, r2, r3
	})
}

func (f *ClientListOrgRepositoriesFunc) nextHook() func(context.Context, string, int, string) ([]*github.Repository, bool, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListOrgRepositoriesFunc) bppendCbll(r0 ClientListOrgRepositoriesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListOrgRepositoriesFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientListOrgRepositoriesFunc) History() []ClientListOrgRepositoriesFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListOrgRepositoriesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListOrgRepositoriesFuncCbll is bn object thbt describes bn
// invocbtion of method ListOrgRepositories on bn instbnce of MockClient.
type ClientListOrgRepositoriesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 int
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListOrgRepositoriesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListOrgRepositoriesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// ClientListOrgbnizbtionMembersFunc describes the behbvior when the
// ListOrgbnizbtionMembers method of the pbrent MockClient instbnce is
// invoked.
type ClientListOrgbnizbtionMembersFunc struct {
	defbultHook func(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error)
	hooks       []func(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error)
	history     []ClientListOrgbnizbtionMembersFuncCbll
	mutex       sync.Mutex
}

// ListOrgbnizbtionMembers delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListOrgbnizbtionMembers(v0 context.Context, v1 string, v2 int, v3 bool) ([]*github.Collbborbtor, bool, error) {
	r0, r1, r2 := m.ListOrgbnizbtionMembersFunc.nextHook()(v0, v1, v2, v3)
	m.ListOrgbnizbtionMembersFunc.bppendCbll(ClientListOrgbnizbtionMembersFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ListOrgbnizbtionMembers method of the pbrent MockClient instbnce is
// invoked bnd the hook queue is empty.
func (f *ClientListOrgbnizbtionMembersFunc) SetDefbultHook(hook func(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListOrgbnizbtionMembers method of the pbrent MockClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ClientListOrgbnizbtionMembersFunc) PushHook(hook func(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListOrgbnizbtionMembersFunc) SetDefbultReturn(r0 []*github.Collbborbtor, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListOrgbnizbtionMembersFunc) PushReturn(r0 []*github.Collbborbtor, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error) {
		return r0, r1, r2
	})
}

func (f *ClientListOrgbnizbtionMembersFunc) nextHook() func(context.Context, string, int, bool) ([]*github.Collbborbtor, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListOrgbnizbtionMembersFunc) bppendCbll(r0 ClientListOrgbnizbtionMembersFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListOrgbnizbtionMembersFuncCbll
// objects describing the invocbtions of this function.
func (f *ClientListOrgbnizbtionMembersFunc) History() []ClientListOrgbnizbtionMembersFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListOrgbnizbtionMembersFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListOrgbnizbtionMembersFuncCbll is bn object thbt describes bn
// invocbtion of method ListOrgbnizbtionMembers on bn instbnce of
// MockClient.
type ClientListOrgbnizbtionMembersFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Collbborbtor
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListOrgbnizbtionMembersFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListOrgbnizbtionMembersFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// ClientListRepositoryCollbborbtorsFunc describes the behbvior when the
// ListRepositoryCollbborbtors method of the pbrent MockClient instbnce is
// invoked.
type ClientListRepositoryCollbborbtorsFunc struct {
	defbultHook func(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error)
	hooks       []func(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error)
	history     []ClientListRepositoryCollbborbtorsFuncCbll
	mutex       sync.Mutex
}

// ListRepositoryCollbborbtors delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListRepositoryCollbborbtors(v0 context.Context, v1 string, v2 string, v3 int, v4 github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error) {
	r0, r1, r2 := m.ListRepositoryCollbborbtorsFunc.nextHook()(v0, v1, v2, v3, v4)
	m.ListRepositoryCollbborbtorsFunc.bppendCbll(ClientListRepositoryCollbborbtorsFuncCbll{v0, v1, v2, v3, v4, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ListRepositoryCollbborbtors method of the pbrent MockClient instbnce is
// invoked bnd the hook queue is empty.
func (f *ClientListRepositoryCollbborbtorsFunc) SetDefbultHook(hook func(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListRepositoryCollbborbtors method of the pbrent MockClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ClientListRepositoryCollbborbtorsFunc) PushHook(hook func(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListRepositoryCollbborbtorsFunc) SetDefbultReturn(r0 []*github.Collbborbtor, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListRepositoryCollbborbtorsFunc) PushReturn(r0 []*github.Collbborbtor, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error) {
		return r0, r1, r2
	})
}

func (f *ClientListRepositoryCollbborbtorsFunc) nextHook() func(context.Context, string, string, int, github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListRepositoryCollbborbtorsFunc) bppendCbll(r0 ClientListRepositoryCollbborbtorsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListRepositoryCollbborbtorsFuncCbll
// objects describing the invocbtions of this function.
func (f *ClientListRepositoryCollbborbtorsFunc) History() []ClientListRepositoryCollbborbtorsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListRepositoryCollbborbtorsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListRepositoryCollbborbtorsFuncCbll is bn object thbt describes bn
// invocbtion of method ListRepositoryCollbborbtors on bn instbnce of
// MockClient.
type ClientListRepositoryCollbborbtorsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 github.CollbborbtorAffilibtion
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Collbborbtor
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListRepositoryCollbborbtorsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListRepositoryCollbborbtorsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// ClientListRepositoryTebmsFunc describes the behbvior when the
// ListRepositoryTebms method of the pbrent MockClient instbnce is invoked.
type ClientListRepositoryTebmsFunc struct {
	defbultHook func(context.Context, string, string, int) ([]*github.Tebm, bool, error)
	hooks       []func(context.Context, string, string, int) ([]*github.Tebm, bool, error)
	history     []ClientListRepositoryTebmsFuncCbll
	mutex       sync.Mutex
}

// ListRepositoryTebms delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListRepositoryTebms(v0 context.Context, v1 string, v2 string, v3 int) ([]*github.Tebm, bool, error) {
	r0, r1, r2 := m.ListRepositoryTebmsFunc.nextHook()(v0, v1, v2, v3)
	m.ListRepositoryTebmsFunc.bppendCbll(ClientListRepositoryTebmsFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the ListRepositoryTebms
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientListRepositoryTebmsFunc) SetDefbultHook(hook func(context.Context, string, string, int) ([]*github.Tebm, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListRepositoryTebms method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientListRepositoryTebmsFunc) PushHook(hook func(context.Context, string, string, int) ([]*github.Tebm, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListRepositoryTebmsFunc) SetDefbultReturn(r0 []*github.Tebm, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, string, int) ([]*github.Tebm, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListRepositoryTebmsFunc) PushReturn(r0 []*github.Tebm, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, string, int) ([]*github.Tebm, bool, error) {
		return r0, r1, r2
	})
}

func (f *ClientListRepositoryTebmsFunc) nextHook() func(context.Context, string, string, int) ([]*github.Tebm, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListRepositoryTebmsFunc) bppendCbll(r0 ClientListRepositoryTebmsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListRepositoryTebmsFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientListRepositoryTebmsFunc) History() []ClientListRepositoryTebmsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListRepositoryTebmsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListRepositoryTebmsFuncCbll is bn object thbt describes bn
// invocbtion of method ListRepositoryTebms on bn instbnce of MockClient.
type ClientListRepositoryTebmsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Tebm
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListRepositoryTebmsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListRepositoryTebmsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// ClientListTebmMembersFunc describes the behbvior when the ListTebmMembers
// method of the pbrent MockClient instbnce is invoked.
type ClientListTebmMembersFunc struct {
	defbultHook func(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error)
	hooks       []func(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error)
	history     []ClientListTebmMembersFuncCbll
	mutex       sync.Mutex
}

// ListTebmMembers delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListTebmMembers(v0 context.Context, v1 string, v2 string, v3 int) ([]*github.Collbborbtor, bool, error) {
	r0, r1, r2 := m.ListTebmMembersFunc.nextHook()(v0, v1, v2, v3)
	m.ListTebmMembersFunc.bppendCbll(ClientListTebmMembersFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the ListTebmMembers
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientListTebmMembersFunc) SetDefbultHook(hook func(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListTebmMembers method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientListTebmMembersFunc) PushHook(hook func(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListTebmMembersFunc) SetDefbultReturn(r0 []*github.Collbborbtor, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListTebmMembersFunc) PushReturn(r0 []*github.Collbborbtor, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error) {
		return r0, r1, r2
	})
}

func (f *ClientListTebmMembersFunc) nextHook() func(context.Context, string, string, int) ([]*github.Collbborbtor, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListTebmMembersFunc) bppendCbll(r0 ClientListTebmMembersFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListTebmMembersFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientListTebmMembersFunc) History() []ClientListTebmMembersFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListTebmMembersFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListTebmMembersFuncCbll is bn object thbt describes bn invocbtion
// of method ListTebmMembers on bn instbnce of MockClient.
type ClientListTebmMembersFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Collbborbtor
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListTebmMembersFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListTebmMembersFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// ClientListTebmRepositoriesFunc describes the behbvior when the
// ListTebmRepositories method of the pbrent MockClient instbnce is invoked.
type ClientListTebmRepositoriesFunc struct {
	defbultHook func(context.Context, string, string, int) ([]*github.Repository, bool, int, error)
	hooks       []func(context.Context, string, string, int) ([]*github.Repository, bool, int, error)
	history     []ClientListTebmRepositoriesFuncCbll
	mutex       sync.Mutex
}

// ListTebmRepositories delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListTebmRepositories(v0 context.Context, v1 string, v2 string, v3 int) ([]*github.Repository, bool, int, error) {
	r0, r1, r2, r3 := m.ListTebmRepositoriesFunc.nextHook()(v0, v1, v2, v3)
	m.ListTebmRepositoriesFunc.bppendCbll(ClientListTebmRepositoriesFuncCbll{v0, v1, v2, v3, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the ListTebmRepositories
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientListTebmRepositoriesFunc) SetDefbultHook(hook func(context.Context, string, string, int) ([]*github.Repository, bool, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListTebmRepositories method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientListTebmRepositoriesFunc) PushHook(hook func(context.Context, string, string, int) ([]*github.Repository, bool, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListTebmRepositoriesFunc) SetDefbultReturn(r0 []*github.Repository, r1 bool, r2 int, r3 error) {
	f.SetDefbultHook(func(context.Context, string, string, int) ([]*github.Repository, bool, int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListTebmRepositoriesFunc) PushReturn(r0 []*github.Repository, r1 bool, r2 int, r3 error) {
	f.PushHook(func(context.Context, string, string, int) ([]*github.Repository, bool, int, error) {
		return r0, r1, r2, r3
	})
}

func (f *ClientListTebmRepositoriesFunc) nextHook() func(context.Context, string, string, int) ([]*github.Repository, bool, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListTebmRepositoriesFunc) bppendCbll(r0 ClientListTebmRepositoriesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListTebmRepositoriesFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientListTebmRepositoriesFunc) History() []ClientListTebmRepositoriesFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListTebmRepositoriesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListTebmRepositoriesFuncCbll is bn object thbt describes bn
// invocbtion of method ListTebmRepositories on bn instbnce of MockClient.
type ClientListTebmRepositoriesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 int
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListTebmRepositoriesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListTebmRepositoriesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// ClientSetWbitForRbteLimitFunc describes the behbvior when the
// SetWbitForRbteLimit method of the pbrent MockClient instbnce is invoked.
type ClientSetWbitForRbteLimitFunc struct {
	defbultHook func(bool)
	hooks       []func(bool)
	history     []ClientSetWbitForRbteLimitFuncCbll
	mutex       sync.Mutex
}

// SetWbitForRbteLimit delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) SetWbitForRbteLimit(v0 bool) {
	m.SetWbitForRbteLimitFunc.nextHook()(v0)
	m.SetWbitForRbteLimitFunc.bppendCbll(ClientSetWbitForRbteLimitFuncCbll{v0})
	return
}

// SetDefbultHook sets function thbt is cblled when the SetWbitForRbteLimit
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientSetWbitForRbteLimitFunc) SetDefbultHook(hook func(bool)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetWbitForRbteLimit method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientSetWbitForRbteLimitFunc) PushHook(hook func(bool)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientSetWbitForRbteLimitFunc) SetDefbultReturn() {
	f.SetDefbultHook(func(bool) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientSetWbitForRbteLimitFunc) PushReturn() {
	f.PushHook(func(bool) {
		return
	})
}

func (f *ClientSetWbitForRbteLimitFunc) nextHook() func(bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientSetWbitForRbteLimitFunc) bppendCbll(r0 ClientSetWbitForRbteLimitFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientSetWbitForRbteLimitFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientSetWbitForRbteLimitFunc) History() []ClientSetWbitForRbteLimitFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientSetWbitForRbteLimitFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientSetWbitForRbteLimitFuncCbll is bn object thbt describes bn
// invocbtion of method SetWbitForRbteLimit on bn instbnce of MockClient.
type ClientSetWbitForRbteLimitFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 bool
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientSetWbitForRbteLimitFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientSetWbitForRbteLimitFuncCbll) Results() []interfbce{} {
	return []interfbce{}{}
}

// ClientWithAuthenticbtorFunc describes the behbvior when the
// WithAuthenticbtor method of the pbrent MockClient instbnce is invoked.
type ClientWithAuthenticbtorFunc struct {
	defbultHook func(buth.Authenticbtor) client
	hooks       []func(buth.Authenticbtor) client
	history     []ClientWithAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// WithAuthenticbtor delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) WithAuthenticbtor(v0 buth.Authenticbtor) client {
	r0 := m.WithAuthenticbtorFunc.nextHook()(v0)
	m.WithAuthenticbtorFunc.bppendCbll(ClientWithAuthenticbtorFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithAuthenticbtor
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientWithAuthenticbtorFunc) SetDefbultHook(hook func(buth.Authenticbtor) client) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithAuthenticbtor method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientWithAuthenticbtorFunc) PushHook(hook func(buth.Authenticbtor) client) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientWithAuthenticbtorFunc) SetDefbultReturn(r0 client) {
	f.SetDefbultHook(func(buth.Authenticbtor) client {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientWithAuthenticbtorFunc) PushReturn(r0 client) {
	f.PushHook(func(buth.Authenticbtor) client {
		return r0
	})
}

func (f *ClientWithAuthenticbtorFunc) nextHook() func(buth.Authenticbtor) client {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientWithAuthenticbtorFunc) bppendCbll(r0 ClientWithAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientWithAuthenticbtorFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientWithAuthenticbtorFunc) History() []ClientWithAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientWithAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientWithAuthenticbtorFuncCbll is bn object thbt describes bn invocbtion
// of method WithAuthenticbtor on bn instbnce of MockClient.
type ClientWithAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 buth.Authenticbtor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 client
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientWithAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientWithAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
