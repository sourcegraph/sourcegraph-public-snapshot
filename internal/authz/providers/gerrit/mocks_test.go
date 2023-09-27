// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge gerrit

import (
	"context"
	"net/url"
	"sync"

	buth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	gerrit "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
)

// MockGerritClient is b mock implementbtion of the Client interfbce (from
// the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit)
// used for unit testing.
type MockGerritClient struct {
	// AbbndonChbngeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AbbndonChbnge.
	AbbndonChbngeFunc *GerritClientAbbndonChbngeFunc
	// AuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method Authenticbtor.
	AuthenticbtorFunc *GerritClientAuthenticbtorFunc
	// DeleteChbngeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method DeleteChbnge.
	DeleteChbngeFunc *GerritClientDeleteChbngeFunc
	// GetAuthenticbtedUserAccountFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetAuthenticbtedUserAccount.
	GetAuthenticbtedUserAccountFunc *GerritClientGetAuthenticbtedUserAccountFunc
	// GetChbngeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetChbnge.
	GetChbngeFunc *GerritClientGetChbngeFunc
	// GetChbngeReviewsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetChbngeReviews.
	GetChbngeReviewsFunc *GerritClientGetChbngeReviewsFunc
	// GetGroupFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetGroup.
	GetGroupFunc *GerritClientGetGroupFunc
	// GetURLFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetURL.
	GetURLFunc *GerritClientGetURLFunc
	// ListProjectsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ListProjects.
	ListProjectsFunc *GerritClientListProjectsFunc
	// MoveChbngeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MoveChbnge.
	MoveChbngeFunc *GerritClientMoveChbngeFunc
	// RestoreChbngeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RestoreChbnge.
	RestoreChbngeFunc *GerritClientRestoreChbngeFunc
	// SetCommitMessbgeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetCommitMessbge.
	SetCommitMessbgeFunc *GerritClientSetCommitMessbgeFunc
	// SetRebdyForReviewFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetRebdyForReview.
	SetRebdyForReviewFunc *GerritClientSetRebdyForReviewFunc
	// SetWIPFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method SetWIP.
	SetWIPFunc *GerritClientSetWIPFunc
	// SubmitChbngeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SubmitChbnge.
	SubmitChbngeFunc *GerritClientSubmitChbngeFunc
	// WithAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithAuthenticbtor.
	WithAuthenticbtorFunc *GerritClientWithAuthenticbtorFunc
	// WriteReviewCommentFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WriteReviewComment.
	WriteReviewCommentFunc *GerritClientWriteReviewCommentFunc
}

// NewMockGerritClient crebtes b new mock of the Client interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockGerritClient() *MockGerritClient {
	return &MockGerritClient{
		AbbndonChbngeFunc: &GerritClientAbbndonChbngeFunc{
			defbultHook: func(context.Context, string) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		AuthenticbtorFunc: &GerritClientAuthenticbtorFunc{
			defbultHook: func() (r0 buth.Authenticbtor) {
				return
			},
		},
		DeleteChbngeFunc: &GerritClientDeleteChbngeFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		GetAuthenticbtedUserAccountFunc: &GerritClientGetAuthenticbtedUserAccountFunc{
			defbultHook: func(context.Context) (r0 *gerrit.Account, r1 error) {
				return
			},
		},
		GetChbngeFunc: &GerritClientGetChbngeFunc{
			defbultHook: func(context.Context, string) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		GetChbngeReviewsFunc: &GerritClientGetChbngeReviewsFunc{
			defbultHook: func(context.Context, string) (r0 *[]gerrit.Reviewer, r1 error) {
				return
			},
		},
		GetGroupFunc: &GerritClientGetGroupFunc{
			defbultHook: func(context.Context, string) (r0 gerrit.Group, r1 error) {
				return
			},
		},
		GetURLFunc: &GerritClientGetURLFunc{
			defbultHook: func() (r0 *url.URL) {
				return
			},
		},
		ListProjectsFunc: &GerritClientListProjectsFunc{
			defbultHook: func(context.Context, gerrit.ListProjectsArgs) (r0 gerrit.ListProjectsResponse, r1 bool, r2 error) {
				return
			},
		},
		MoveChbngeFunc: &GerritClientMoveChbngeFunc{
			defbultHook: func(context.Context, string, gerrit.MoveChbngePbylobd) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		RestoreChbngeFunc: &GerritClientRestoreChbngeFunc{
			defbultHook: func(context.Context, string) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		SetCommitMessbgeFunc: &GerritClientSetCommitMessbgeFunc{
			defbultHook: func(context.Context, string, gerrit.SetCommitMessbgePbylobd) (r0 error) {
				return
			},
		},
		SetRebdyForReviewFunc: &GerritClientSetRebdyForReviewFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		SetWIPFunc: &GerritClientSetWIPFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		SubmitChbngeFunc: &GerritClientSubmitChbngeFunc{
			defbultHook: func(context.Context, string) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		WithAuthenticbtorFunc: &GerritClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (r0 gerrit.Client, r1 error) {
				return
			},
		},
		WriteReviewCommentFunc: &GerritClientWriteReviewCommentFunc{
			defbultHook: func(context.Context, string, gerrit.ChbngeReviewComment) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockGerritClient crebtes b new mock of the Client interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGerritClient() *MockGerritClient {
	return &MockGerritClient{
		AbbndonChbngeFunc: &GerritClientAbbndonChbngeFunc{
			defbultHook: func(context.Context, string) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.AbbndonChbnge")
			},
		},
		AuthenticbtorFunc: &GerritClientAuthenticbtorFunc{
			defbultHook: func() buth.Authenticbtor {
				pbnic("unexpected invocbtion of MockGerritClient.Authenticbtor")
			},
		},
		DeleteChbngeFunc: &GerritClientDeleteChbngeFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockGerritClient.DeleteChbnge")
			},
		},
		GetAuthenticbtedUserAccountFunc: &GerritClientGetAuthenticbtedUserAccountFunc{
			defbultHook: func(context.Context) (*gerrit.Account, error) {
				pbnic("unexpected invocbtion of MockGerritClient.GetAuthenticbtedUserAccount")
			},
		},
		GetChbngeFunc: &GerritClientGetChbngeFunc{
			defbultHook: func(context.Context, string) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.GetChbnge")
			},
		},
		GetChbngeReviewsFunc: &GerritClientGetChbngeReviewsFunc{
			defbultHook: func(context.Context, string) (*[]gerrit.Reviewer, error) {
				pbnic("unexpected invocbtion of MockGerritClient.GetChbngeReviews")
			},
		},
		GetGroupFunc: &GerritClientGetGroupFunc{
			defbultHook: func(context.Context, string) (gerrit.Group, error) {
				pbnic("unexpected invocbtion of MockGerritClient.GetGroup")
			},
		},
		GetURLFunc: &GerritClientGetURLFunc{
			defbultHook: func() *url.URL {
				pbnic("unexpected invocbtion of MockGerritClient.GetURL")
			},
		},
		ListProjectsFunc: &GerritClientListProjectsFunc{
			defbultHook: func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
				pbnic("unexpected invocbtion of MockGerritClient.ListProjects")
			},
		},
		MoveChbngeFunc: &GerritClientMoveChbngeFunc{
			defbultHook: func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.MoveChbnge")
			},
		},
		RestoreChbngeFunc: &GerritClientRestoreChbngeFunc{
			defbultHook: func(context.Context, string) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.RestoreChbnge")
			},
		},
		SetCommitMessbgeFunc: &GerritClientSetCommitMessbgeFunc{
			defbultHook: func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error {
				pbnic("unexpected invocbtion of MockGerritClient.SetCommitMessbge")
			},
		},
		SetRebdyForReviewFunc: &GerritClientSetRebdyForReviewFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockGerritClient.SetRebdyForReview")
			},
		},
		SetWIPFunc: &GerritClientSetWIPFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockGerritClient.SetWIP")
			},
		},
		SubmitChbngeFunc: &GerritClientSubmitChbngeFunc{
			defbultHook: func(context.Context, string) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.SubmitChbnge")
			},
		},
		WithAuthenticbtorFunc: &GerritClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (gerrit.Client, error) {
				pbnic("unexpected invocbtion of MockGerritClient.WithAuthenticbtor")
			},
		},
		WriteReviewCommentFunc: &GerritClientWriteReviewCommentFunc{
			defbultHook: func(context.Context, string, gerrit.ChbngeReviewComment) error {
				pbnic("unexpected invocbtion of MockGerritClient.WriteReviewComment")
			},
		},
	}
}

// NewMockGerritClientFrom crebtes b new mock of the MockGerritClient
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGerritClientFrom(i gerrit.Client) *MockGerritClient {
	return &MockGerritClient{
		AbbndonChbngeFunc: &GerritClientAbbndonChbngeFunc{
			defbultHook: i.AbbndonChbnge,
		},
		AuthenticbtorFunc: &GerritClientAuthenticbtorFunc{
			defbultHook: i.Authenticbtor,
		},
		DeleteChbngeFunc: &GerritClientDeleteChbngeFunc{
			defbultHook: i.DeleteChbnge,
		},
		GetAuthenticbtedUserAccountFunc: &GerritClientGetAuthenticbtedUserAccountFunc{
			defbultHook: i.GetAuthenticbtedUserAccount,
		},
		GetChbngeFunc: &GerritClientGetChbngeFunc{
			defbultHook: i.GetChbnge,
		},
		GetChbngeReviewsFunc: &GerritClientGetChbngeReviewsFunc{
			defbultHook: i.GetChbngeReviews,
		},
		GetGroupFunc: &GerritClientGetGroupFunc{
			defbultHook: i.GetGroup,
		},
		GetURLFunc: &GerritClientGetURLFunc{
			defbultHook: i.GetURL,
		},
		ListProjectsFunc: &GerritClientListProjectsFunc{
			defbultHook: i.ListProjects,
		},
		MoveChbngeFunc: &GerritClientMoveChbngeFunc{
			defbultHook: i.MoveChbnge,
		},
		RestoreChbngeFunc: &GerritClientRestoreChbngeFunc{
			defbultHook: i.RestoreChbnge,
		},
		SetCommitMessbgeFunc: &GerritClientSetCommitMessbgeFunc{
			defbultHook: i.SetCommitMessbge,
		},
		SetRebdyForReviewFunc: &GerritClientSetRebdyForReviewFunc{
			defbultHook: i.SetRebdyForReview,
		},
		SetWIPFunc: &GerritClientSetWIPFunc{
			defbultHook: i.SetWIP,
		},
		SubmitChbngeFunc: &GerritClientSubmitChbngeFunc{
			defbultHook: i.SubmitChbnge,
		},
		WithAuthenticbtorFunc: &GerritClientWithAuthenticbtorFunc{
			defbultHook: i.WithAuthenticbtor,
		},
		WriteReviewCommentFunc: &GerritClientWriteReviewCommentFunc{
			defbultHook: i.WriteReviewComment,
		},
	}
}

// GerritClientAbbndonChbngeFunc describes the behbvior when the
// AbbndonChbnge method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientAbbndonChbngeFunc struct {
	defbultHook func(context.Context, string) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string) (*gerrit.Chbnge, error)
	history     []GerritClientAbbndonChbngeFuncCbll
	mutex       sync.Mutex
}

// AbbndonChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) AbbndonChbnge(v0 context.Context, v1 string) (*gerrit.Chbnge, error) {
	r0, r1 := m.AbbndonChbngeFunc.nextHook()(v0, v1)
	m.AbbndonChbngeFunc.bppendCbll(GerritClientAbbndonChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the AbbndonChbnge method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientAbbndonChbngeFunc) SetDefbultHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AbbndonChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientAbbndonChbngeFunc) PushHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientAbbndonChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientAbbndonChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientAbbndonChbngeFunc) nextHook() func(context.Context, string) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientAbbndonChbngeFunc) bppendCbll(r0 GerritClientAbbndonChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientAbbndonChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientAbbndonChbngeFunc) History() []GerritClientAbbndonChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientAbbndonChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientAbbndonChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method AbbndonChbnge on bn instbnce of MockGerritClient.
type GerritClientAbbndonChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientAbbndonChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientAbbndonChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientAuthenticbtorFunc describes the behbvior when the
// Authenticbtor method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientAuthenticbtorFunc struct {
	defbultHook func() buth.Authenticbtor
	hooks       []func() buth.Authenticbtor
	history     []GerritClientAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// Authenticbtor delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) Authenticbtor() buth.Authenticbtor {
	r0 := m.AuthenticbtorFunc.nextHook()()
	m.AuthenticbtorFunc.bppendCbll(GerritClientAuthenticbtorFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Authenticbtor method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientAuthenticbtorFunc) SetDefbultHook(hook func() buth.Authenticbtor) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Authenticbtor method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientAuthenticbtorFunc) PushHook(hook func() buth.Authenticbtor) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientAuthenticbtorFunc) SetDefbultReturn(r0 buth.Authenticbtor) {
	f.SetDefbultHook(func() buth.Authenticbtor {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientAuthenticbtorFunc) PushReturn(r0 buth.Authenticbtor) {
	f.PushHook(func() buth.Authenticbtor {
		return r0
	})
}

func (f *GerritClientAuthenticbtorFunc) nextHook() func() buth.Authenticbtor {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientAuthenticbtorFunc) bppendCbll(r0 GerritClientAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientAuthenticbtorFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientAuthenticbtorFunc) History() []GerritClientAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientAuthenticbtorFuncCbll is bn object thbt describes bn
// invocbtion of method Authenticbtor on bn instbnce of MockGerritClient.
type GerritClientAuthenticbtorFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 buth.Authenticbtor
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientDeleteChbngeFunc describes the behbvior when the DeleteChbnge
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientDeleteChbngeFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []GerritClientDeleteChbngeFuncCbll
	mutex       sync.Mutex
}

// DeleteChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) DeleteChbnge(v0 context.Context, v1 string) error {
	r0 := m.DeleteChbngeFunc.nextHook()(v0, v1)
	m.DeleteChbngeFunc.bppendCbll(GerritClientDeleteChbngeFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the DeleteChbnge method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientDeleteChbngeFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientDeleteChbngeFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientDeleteChbngeFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientDeleteChbngeFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *GerritClientDeleteChbngeFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientDeleteChbngeFunc) bppendCbll(r0 GerritClientDeleteChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientDeleteChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientDeleteChbngeFunc) History() []GerritClientDeleteChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientDeleteChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientDeleteChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteChbnge on bn instbnce of MockGerritClient.
type GerritClientDeleteChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientDeleteChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientDeleteChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientGetAuthenticbtedUserAccountFunc describes the behbvior when
// the GetAuthenticbtedUserAccount method of the pbrent MockGerritClient
// instbnce is invoked.
type GerritClientGetAuthenticbtedUserAccountFunc struct {
	defbultHook func(context.Context) (*gerrit.Account, error)
	hooks       []func(context.Context) (*gerrit.Account, error)
	history     []GerritClientGetAuthenticbtedUserAccountFuncCbll
	mutex       sync.Mutex
}

// GetAuthenticbtedUserAccount delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetAuthenticbtedUserAccount(v0 context.Context) (*gerrit.Account, error) {
	r0, r1 := m.GetAuthenticbtedUserAccountFunc.nextHook()(v0)
	m.GetAuthenticbtedUserAccountFunc.bppendCbll(GerritClientGetAuthenticbtedUserAccountFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetAuthenticbtedUserAccount method of the pbrent MockGerritClient
// instbnce is invoked bnd the hook queue is empty.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) SetDefbultHook(hook func(context.Context) (*gerrit.Account, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAuthenticbtedUserAccount method of the pbrent MockGerritClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) PushHook(hook func(context.Context) (*gerrit.Account, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) SetDefbultReturn(r0 *gerrit.Account, r1 error) {
	f.SetDefbultHook(func(context.Context) (*gerrit.Account, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) PushReturn(r0 *gerrit.Account, r1 error) {
	f.PushHook(func(context.Context) (*gerrit.Account, error) {
		return r0, r1
	})
}

func (f *GerritClientGetAuthenticbtedUserAccountFunc) nextHook() func(context.Context) (*gerrit.Account, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetAuthenticbtedUserAccountFunc) bppendCbll(r0 GerritClientGetAuthenticbtedUserAccountFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GerritClientGetAuthenticbtedUserAccountFuncCbll objects describing the
// invocbtions of this function.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) History() []GerritClientGetAuthenticbtedUserAccountFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetAuthenticbtedUserAccountFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetAuthenticbtedUserAccountFuncCbll is bn object thbt
// describes bn invocbtion of method GetAuthenticbtedUserAccount on bn
// instbnce of MockGerritClient.
type GerritClientGetAuthenticbtedUserAccountFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Account
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetAuthenticbtedUserAccountFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetAuthenticbtedUserAccountFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientGetChbngeFunc describes the behbvior when the GetChbnge
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientGetChbngeFunc struct {
	defbultHook func(context.Context, string) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string) (*gerrit.Chbnge, error)
	history     []GerritClientGetChbngeFuncCbll
	mutex       sync.Mutex
}

// GetChbnge delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetChbnge(v0 context.Context, v1 string) (*gerrit.Chbnge, error) {
	r0, r1 := m.GetChbngeFunc.nextHook()(v0, v1)
	m.GetChbngeFunc.bppendCbll(GerritClientGetChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetChbnge method of
// the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientGetChbngeFunc) SetDefbultHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetChbnge method of the pbrent MockGerritClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GerritClientGetChbngeFunc) PushHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientGetChbngeFunc) nextHook() func(context.Context, string) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetChbngeFunc) bppendCbll(r0 GerritClientGetChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientGetChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientGetChbngeFunc) History() []GerritClientGetChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetChbngeFuncCbll is bn object thbt describes bn invocbtion
// of method GetChbnge on bn instbnce of MockGerritClient.
type GerritClientGetChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientGetChbngeReviewsFunc describes the behbvior when the
// GetChbngeReviews method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientGetChbngeReviewsFunc struct {
	defbultHook func(context.Context, string) (*[]gerrit.Reviewer, error)
	hooks       []func(context.Context, string) (*[]gerrit.Reviewer, error)
	history     []GerritClientGetChbngeReviewsFuncCbll
	mutex       sync.Mutex
}

// GetChbngeReviews delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetChbngeReviews(v0 context.Context, v1 string) (*[]gerrit.Reviewer, error) {
	r0, r1 := m.GetChbngeReviewsFunc.nextHook()(v0, v1)
	m.GetChbngeReviewsFunc.bppendCbll(GerritClientGetChbngeReviewsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetChbngeReviews
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientGetChbngeReviewsFunc) SetDefbultHook(hook func(context.Context, string) (*[]gerrit.Reviewer, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetChbngeReviews method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientGetChbngeReviewsFunc) PushHook(hook func(context.Context, string) (*[]gerrit.Reviewer, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetChbngeReviewsFunc) SetDefbultReturn(r0 *[]gerrit.Reviewer, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*[]gerrit.Reviewer, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetChbngeReviewsFunc) PushReturn(r0 *[]gerrit.Reviewer, r1 error) {
	f.PushHook(func(context.Context, string) (*[]gerrit.Reviewer, error) {
		return r0, r1
	})
}

func (f *GerritClientGetChbngeReviewsFunc) nextHook() func(context.Context, string) (*[]gerrit.Reviewer, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetChbngeReviewsFunc) bppendCbll(r0 GerritClientGetChbngeReviewsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientGetChbngeReviewsFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientGetChbngeReviewsFunc) History() []GerritClientGetChbngeReviewsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetChbngeReviewsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetChbngeReviewsFuncCbll is bn object thbt describes bn
// invocbtion of method GetChbngeReviews on bn instbnce of MockGerritClient.
type GerritClientGetChbngeReviewsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *[]gerrit.Reviewer
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetChbngeReviewsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetChbngeReviewsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientGetGroupFunc describes the behbvior when the GetGroup method
// of the pbrent MockGerritClient instbnce is invoked.
type GerritClientGetGroupFunc struct {
	defbultHook func(context.Context, string) (gerrit.Group, error)
	hooks       []func(context.Context, string) (gerrit.Group, error)
	history     []GerritClientGetGroupFuncCbll
	mutex       sync.Mutex
}

// GetGroup delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetGroup(v0 context.Context, v1 string) (gerrit.Group, error) {
	r0, r1 := m.GetGroupFunc.nextHook()(v0, v1)
	m.GetGroupFunc.bppendCbll(GerritClientGetGroupFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetGroup method of
// the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientGetGroupFunc) SetDefbultHook(hook func(context.Context, string) (gerrit.Group, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetGroup method of the pbrent MockGerritClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GerritClientGetGroupFunc) PushHook(hook func(context.Context, string) (gerrit.Group, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetGroupFunc) SetDefbultReturn(r0 gerrit.Group, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (gerrit.Group, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetGroupFunc) PushReturn(r0 gerrit.Group, r1 error) {
	f.PushHook(func(context.Context, string) (gerrit.Group, error) {
		return r0, r1
	})
}

func (f *GerritClientGetGroupFunc) nextHook() func(context.Context, string) (gerrit.Group, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetGroupFunc) bppendCbll(r0 GerritClientGetGroupFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientGetGroupFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientGetGroupFunc) History() []GerritClientGetGroupFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetGroupFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetGroupFuncCbll is bn object thbt describes bn invocbtion of
// method GetGroup on bn instbnce of MockGerritClient.
type GerritClientGetGroupFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gerrit.Group
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetGroupFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetGroupFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientGetURLFunc describes the behbvior when the GetURL method of
// the pbrent MockGerritClient instbnce is invoked.
type GerritClientGetURLFunc struct {
	defbultHook func() *url.URL
	hooks       []func() *url.URL
	history     []GerritClientGetURLFuncCbll
	mutex       sync.Mutex
}

// GetURL delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetURL() *url.URL {
	r0 := m.GetURLFunc.nextHook()()
	m.GetURLFunc.bppendCbll(GerritClientGetURLFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the GetURL method of the
// pbrent MockGerritClient instbnce is invoked bnd the hook queue is empty.
func (f *GerritClientGetURLFunc) SetDefbultHook(hook func() *url.URL) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetURL method of the pbrent MockGerritClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GerritClientGetURLFunc) PushHook(hook func() *url.URL) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetURLFunc) SetDefbultReturn(r0 *url.URL) {
	f.SetDefbultHook(func() *url.URL {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetURLFunc) PushReturn(r0 *url.URL) {
	f.PushHook(func() *url.URL {
		return r0
	})
}

func (f *GerritClientGetURLFunc) nextHook() func() *url.URL {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetURLFunc) bppendCbll(r0 GerritClientGetURLFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientGetURLFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientGetURLFunc) History() []GerritClientGetURLFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetURLFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetURLFuncCbll is bn object thbt describes bn invocbtion of
// method GetURL on bn instbnce of MockGerritClient.
type GerritClientGetURLFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *url.URL
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetURLFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetURLFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientListProjectsFunc describes the behbvior when the ListProjects
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientListProjectsFunc struct {
	defbultHook func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)
	hooks       []func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)
	history     []GerritClientListProjectsFuncCbll
	mutex       sync.Mutex
}

// ListProjects delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) ListProjects(v0 context.Context, v1 gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
	r0, r1, r2 := m.ListProjectsFunc.nextHook()(v0, v1)
	m.ListProjectsFunc.bppendCbll(GerritClientListProjectsFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the ListProjects method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientListProjectsFunc) SetDefbultHook(hook func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListProjects method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientListProjectsFunc) PushHook(hook func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientListProjectsFunc) SetDefbultReturn(r0 gerrit.ListProjectsResponse, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientListProjectsFunc) PushReturn(r0 gerrit.ListProjectsResponse, r1 bool, r2 error) {
	f.PushHook(func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
		return r0, r1, r2
	})
}

func (f *GerritClientListProjectsFunc) nextHook() func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientListProjectsFunc) bppendCbll(r0 GerritClientListProjectsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientListProjectsFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientListProjectsFunc) History() []GerritClientListProjectsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientListProjectsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientListProjectsFuncCbll is bn object thbt describes bn
// invocbtion of method ListProjects on bn instbnce of MockGerritClient.
type GerritClientListProjectsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 gerrit.ListProjectsArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gerrit.ListProjectsResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientListProjectsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientListProjectsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// GerritClientMoveChbngeFunc describes the behbvior when the MoveChbnge
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientMoveChbngeFunc struct {
	defbultHook func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error)
	history     []GerritClientMoveChbngeFuncCbll
	mutex       sync.Mutex
}

// MoveChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) MoveChbnge(v0 context.Context, v1 string, v2 gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
	r0, r1 := m.MoveChbngeFunc.nextHook()(v0, v1, v2)
	m.MoveChbngeFunc.bppendCbll(GerritClientMoveChbngeFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MoveChbnge method of
// the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientMoveChbngeFunc) SetDefbultHook(hook func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MoveChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientMoveChbngeFunc) PushHook(hook func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientMoveChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientMoveChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientMoveChbngeFunc) nextHook() func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientMoveChbngeFunc) bppendCbll(r0 GerritClientMoveChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientMoveChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientMoveChbngeFunc) History() []GerritClientMoveChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientMoveChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientMoveChbngeFuncCbll is bn object thbt describes bn invocbtion
// of method MoveChbnge on bn instbnce of MockGerritClient.
type GerritClientMoveChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gerrit.MoveChbngePbylobd
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientMoveChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientMoveChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientRestoreChbngeFunc describes the behbvior when the
// RestoreChbnge method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientRestoreChbngeFunc struct {
	defbultHook func(context.Context, string) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string) (*gerrit.Chbnge, error)
	history     []GerritClientRestoreChbngeFuncCbll
	mutex       sync.Mutex
}

// RestoreChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) RestoreChbnge(v0 context.Context, v1 string) (*gerrit.Chbnge, error) {
	r0, r1 := m.RestoreChbngeFunc.nextHook()(v0, v1)
	m.RestoreChbngeFunc.bppendCbll(GerritClientRestoreChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RestoreChbnge method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientRestoreChbngeFunc) SetDefbultHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RestoreChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientRestoreChbngeFunc) PushHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientRestoreChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientRestoreChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientRestoreChbngeFunc) nextHook() func(context.Context, string) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientRestoreChbngeFunc) bppendCbll(r0 GerritClientRestoreChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientRestoreChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientRestoreChbngeFunc) History() []GerritClientRestoreChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientRestoreChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientRestoreChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method RestoreChbnge on bn instbnce of MockGerritClient.
type GerritClientRestoreChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientRestoreChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientRestoreChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientSetCommitMessbgeFunc describes the behbvior when the
// SetCommitMessbge method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientSetCommitMessbgeFunc struct {
	defbultHook func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error
	hooks       []func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error
	history     []GerritClientSetCommitMessbgeFuncCbll
	mutex       sync.Mutex
}

// SetCommitMessbge delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) SetCommitMessbge(v0 context.Context, v1 string, v2 gerrit.SetCommitMessbgePbylobd) error {
	r0 := m.SetCommitMessbgeFunc.nextHook()(v0, v1, v2)
	m.SetCommitMessbgeFunc.bppendCbll(GerritClientSetCommitMessbgeFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetCommitMessbge
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientSetCommitMessbgeFunc) SetDefbultHook(hook func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetCommitMessbge method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientSetCommitMessbgeFunc) PushHook(hook func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientSetCommitMessbgeFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientSetCommitMessbgeFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error {
		return r0
	})
}

func (f *GerritClientSetCommitMessbgeFunc) nextHook() func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientSetCommitMessbgeFunc) bppendCbll(r0 GerritClientSetCommitMessbgeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientSetCommitMessbgeFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientSetCommitMessbgeFunc) History() []GerritClientSetCommitMessbgeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientSetCommitMessbgeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientSetCommitMessbgeFuncCbll is bn object thbt describes bn
// invocbtion of method SetCommitMessbge on bn instbnce of MockGerritClient.
type GerritClientSetCommitMessbgeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gerrit.SetCommitMessbgePbylobd
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientSetCommitMessbgeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientSetCommitMessbgeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientSetRebdyForReviewFunc describes the behbvior when the
// SetRebdyForReview method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientSetRebdyForReviewFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []GerritClientSetRebdyForReviewFuncCbll
	mutex       sync.Mutex
}

// SetRebdyForReview delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) SetRebdyForReview(v0 context.Context, v1 string) error {
	r0 := m.SetRebdyForReviewFunc.nextHook()(v0, v1)
	m.SetRebdyForReviewFunc.bppendCbll(GerritClientSetRebdyForReviewFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetRebdyForReview
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientSetRebdyForReviewFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetRebdyForReview method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientSetRebdyForReviewFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientSetRebdyForReviewFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientSetRebdyForReviewFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *GerritClientSetRebdyForReviewFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientSetRebdyForReviewFunc) bppendCbll(r0 GerritClientSetRebdyForReviewFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientSetRebdyForReviewFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientSetRebdyForReviewFunc) History() []GerritClientSetRebdyForReviewFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientSetRebdyForReviewFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientSetRebdyForReviewFuncCbll is bn object thbt describes bn
// invocbtion of method SetRebdyForReview on bn instbnce of
// MockGerritClient.
type GerritClientSetRebdyForReviewFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientSetRebdyForReviewFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientSetRebdyForReviewFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientSetWIPFunc describes the behbvior when the SetWIP method of
// the pbrent MockGerritClient instbnce is invoked.
type GerritClientSetWIPFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []GerritClientSetWIPFuncCbll
	mutex       sync.Mutex
}

// SetWIP delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) SetWIP(v0 context.Context, v1 string) error {
	r0 := m.SetWIPFunc.nextHook()(v0, v1)
	m.SetWIPFunc.bppendCbll(GerritClientSetWIPFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetWIP method of the
// pbrent MockGerritClient instbnce is invoked bnd the hook queue is empty.
func (f *GerritClientSetWIPFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetWIP method of the pbrent MockGerritClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GerritClientSetWIPFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientSetWIPFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientSetWIPFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *GerritClientSetWIPFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientSetWIPFunc) bppendCbll(r0 GerritClientSetWIPFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientSetWIPFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientSetWIPFunc) History() []GerritClientSetWIPFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientSetWIPFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientSetWIPFuncCbll is bn object thbt describes bn invocbtion of
// method SetWIP on bn instbnce of MockGerritClient.
type GerritClientSetWIPFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientSetWIPFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientSetWIPFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientSubmitChbngeFunc describes the behbvior when the SubmitChbnge
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientSubmitChbngeFunc struct {
	defbultHook func(context.Context, string) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string) (*gerrit.Chbnge, error)
	history     []GerritClientSubmitChbngeFuncCbll
	mutex       sync.Mutex
}

// SubmitChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) SubmitChbnge(v0 context.Context, v1 string) (*gerrit.Chbnge, error) {
	r0, r1 := m.SubmitChbngeFunc.nextHook()(v0, v1)
	m.SubmitChbngeFunc.bppendCbll(GerritClientSubmitChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SubmitChbnge method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientSubmitChbngeFunc) SetDefbultHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SubmitChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientSubmitChbngeFunc) PushHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientSubmitChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientSubmitChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientSubmitChbngeFunc) nextHook() func(context.Context, string) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientSubmitChbngeFunc) bppendCbll(r0 GerritClientSubmitChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientSubmitChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientSubmitChbngeFunc) History() []GerritClientSubmitChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientSubmitChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientSubmitChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method SubmitChbnge on bn instbnce of MockGerritClient.
type GerritClientSubmitChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientSubmitChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientSubmitChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientWithAuthenticbtorFunc describes the behbvior when the
// WithAuthenticbtor method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientWithAuthenticbtorFunc struct {
	defbultHook func(buth.Authenticbtor) (gerrit.Client, error)
	hooks       []func(buth.Authenticbtor) (gerrit.Client, error)
	history     []GerritClientWithAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// WithAuthenticbtor delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) WithAuthenticbtor(v0 buth.Authenticbtor) (gerrit.Client, error) {
	r0, r1 := m.WithAuthenticbtorFunc.nextHook()(v0)
	m.WithAuthenticbtorFunc.bppendCbll(GerritClientWithAuthenticbtorFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the WithAuthenticbtor
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientWithAuthenticbtorFunc) SetDefbultHook(hook func(buth.Authenticbtor) (gerrit.Client, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithAuthenticbtor method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientWithAuthenticbtorFunc) PushHook(hook func(buth.Authenticbtor) (gerrit.Client, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientWithAuthenticbtorFunc) SetDefbultReturn(r0 gerrit.Client, r1 error) {
	f.SetDefbultHook(func(buth.Authenticbtor) (gerrit.Client, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientWithAuthenticbtorFunc) PushReturn(r0 gerrit.Client, r1 error) {
	f.PushHook(func(buth.Authenticbtor) (gerrit.Client, error) {
		return r0, r1
	})
}

func (f *GerritClientWithAuthenticbtorFunc) nextHook() func(buth.Authenticbtor) (gerrit.Client, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientWithAuthenticbtorFunc) bppendCbll(r0 GerritClientWithAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientWithAuthenticbtorFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientWithAuthenticbtorFunc) History() []GerritClientWithAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientWithAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientWithAuthenticbtorFuncCbll is bn object thbt describes bn
// invocbtion of method WithAuthenticbtor on bn instbnce of
// MockGerritClient.
type GerritClientWithAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 buth.Authenticbtor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gerrit.Client
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientWithAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientWithAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientWriteReviewCommentFunc describes the behbvior when the
// WriteReviewComment method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientWriteReviewCommentFunc struct {
	defbultHook func(context.Context, string, gerrit.ChbngeReviewComment) error
	hooks       []func(context.Context, string, gerrit.ChbngeReviewComment) error
	history     []GerritClientWriteReviewCommentFuncCbll
	mutex       sync.Mutex
}

// WriteReviewComment delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) WriteReviewComment(v0 context.Context, v1 string, v2 gerrit.ChbngeReviewComment) error {
	r0 := m.WriteReviewCommentFunc.nextHook()(v0, v1, v2)
	m.WriteReviewCommentFunc.bppendCbll(GerritClientWriteReviewCommentFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WriteReviewComment
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientWriteReviewCommentFunc) SetDefbultHook(hook func(context.Context, string, gerrit.ChbngeReviewComment) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WriteReviewComment method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientWriteReviewCommentFunc) PushHook(hook func(context.Context, string, gerrit.ChbngeReviewComment) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientWriteReviewCommentFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, gerrit.ChbngeReviewComment) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientWriteReviewCommentFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, gerrit.ChbngeReviewComment) error {
		return r0
	})
}

func (f *GerritClientWriteReviewCommentFunc) nextHook() func(context.Context, string, gerrit.ChbngeReviewComment) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientWriteReviewCommentFunc) bppendCbll(r0 GerritClientWriteReviewCommentFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientWriteReviewCommentFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientWriteReviewCommentFunc) History() []GerritClientWriteReviewCommentFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientWriteReviewCommentFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientWriteReviewCommentFuncCbll is bn object thbt describes bn
// invocbtion of method WriteReviewComment on bn instbnce of
// MockGerritClient.
type GerritClientWriteReviewCommentFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gerrit.ChbngeReviewComment
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientWriteReviewCommentFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientWriteReviewCommentFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
