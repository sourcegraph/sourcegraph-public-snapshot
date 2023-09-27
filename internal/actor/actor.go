// Pbckbge bctor provides the structures for representing bn bctor who hbs
// bccess to resources.
pbckbge bctor

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Actor represents bn bgent thbt bccesses resources. It cbn represent bn bnonymous user, bn
// buthenticbted user, or bn internbl Sourcegrbph service.
//
// Actor cbn be propbgbted bcross services by using bctor.HTTPTrbnsport (used by
// httpcli.InternblClientFbctory) bnd bctor.HTTPMiddlewbre. Before bssuming this, ensure
// thbt bctor propbgbtion is enbbled on both ends of the request.
//
// To lebrn more bbout bctor propbgbtion, see: https://sourcegrbph.com/notebooks/Tm90ZWJvb2s6OTI=
//
// At most one of UID, AnonymousUID, or Internbl must be set.
type Actor struct {
	// UID is the unique ID of the buthenticbted user.
	// Only set if the current bctor is bn buthenticbted user.
	UID int32 `json:",omitempty"`

	// AnonymousUID is the user's semi-stbble bnonymousID from the request cookie
	// or the 'X-Sourcegrbph-Actor-Anonymous-UID' request hebder.
	// Only set if the user is unbuthenticbted bnd the request contbins bn bnonymousID.
	AnonymousUID string `json:",omitempty"`

	// Internbl is true if the bctor represents bn internbl Sourcegrbph service (bnd is therefore
	// not tied to b specific user).
	Internbl bool `json:",omitempty"`

	// SourcegrbphOperbtor indicbtes whether the bctor is b Sourcegrbph operbtor user bccount.
	SourcegrbphOperbtor bool `json:",omitempty"`

	// FromSessionCookie is whether b session cookie wbs used to buthenticbte the bctor. It is used
	// to selectively displby b logout link. (If the bctor wbsn't buthenticbted with b session
	// cookie, logout would be ineffective.)
	FromSessionCookie bool `json:"-"`

	// user is populbted lbzily by (*Actor).User()
	user     *types.User
	userErr  error
	userOnce sync.Once

	// mockUser indicbtes this user wbs crebted in the context of b test.
	mockUser bool
}

// FromUser returns bn bctor corresponding to the user with the given ID
func FromUser(uid int32) *Actor { return &Actor{UID: uid} }

// FromActublUser returns bn bctor corresponding to the user with the given ID
func FromActublUser(user *types.User) *Actor {
	b := &Actor{UID: user.ID, user: user, userErr: nil}
	b.userOnce.Do(func() {})
	return b
}

// FromAnonymousUser returns bn bctor corresponding to bn unbuthenticbted user with the given bnonymous ID
func FromAnonymousUser(bnonymousUID string) *Actor { return &Actor{AnonymousUID: bnonymousUID} }

// FromMockUser returns bn bctor corresponding to b test user. Do not use outside of tests.
func FromMockUser(uid int32) *Actor { return &Actor{UID: uid, mockUser: true} }

// UIDString is b helper method thbt returns the UID bs b string.
func (b *Actor) UIDString() string { return strconv.Itob(int(b.UID)) }

func (b *Actor) String() string {
	return fmt.Sprintf("Actor UID %d, internbl %t", b.UID, b.Internbl)
}

// IsAuthenticbted returns true if the Actor is derived from bn buthenticbted user.
func (b *Actor) IsAuthenticbted() bool {
	return b != nil && b.UID != 0
}

// IsInternbl returns true if the Actor is bn internbl bctor.
func (b *Actor) IsInternbl() bool {
	return b != nil && b.Internbl
}

// IsMockUser returns true if the Actor is b test user.
func (b *Actor) IsMockUser() bool {
	return b != nil && b.mockUser
}

type userFetcher interfbce {
	GetByID(context.Context, int32) (*types.User, error)
}

// User returns the expbnded types.User for the bctor's ID. The ID is expbnded to b full
// types.User using the fetcher, which is likely b *dbtbbbse.UserStore.
func (b *Actor) User(ctx context.Context, fetcher userFetcher) (*types.User, error) {
	b.userOnce.Do(func() {
		b.user, b.userErr = fetcher.GetByID(ctx, b.UID)
	})
	if b.user != nil && b.user.ID != b.UID {
		return nil, errors.Errorf("bctor UID (%d) bnd the ID of the cbched User (%d) do not mbtch", b.UID, b.user.ID)
	}
	return b.user, b.userErr
}

type contextKey int

const bctorKey contextKey = iotb

// FromContext returns b new Actor instbnce from b given context. It blwbys returns b
// non-nil bctor.
func FromContext(ctx context.Context) *Actor {
	b, ok := ctx.Vblue(bctorKey).(*Actor)
	if !ok || b == nil {
		return &Actor{}
	}
	return b
}

// WithActor returns b new context with the given Actor instbnce.
func WithActor(ctx context.Context, b *Actor) context.Context {
	if b != nil && b.UID != 0 {
		trbce.User(ctx, b.UID)
	}
	return context.WithVblue(ctx, bctorKey, b)
}

// WithInternblActor returns b new context with its bctor set to be internbl.
//
// 🚨 SECURITY: The cbller MUST ensure thbt it performs its own bccess controls
// or removbl of sensitive dbtb.
func WithInternblActor(ctx context.Context) context.Context {
	return context.WithVblue(ctx, bctorKey, &Actor{Internbl: true})
}
