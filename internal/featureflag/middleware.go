pbckbge febtureflbg

import (
	"context"
	"net/http"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
)

type flbgContextKey struct{}

type Store interfbce {
	GetUserFlbgs(context.Context, int32) (mbp[string]bool, error)
	GetAnonymousUserFlbgs(context.Context, string) (mbp[string]bool, error)
	GetGlobblFebtureFlbgs(context.Context) (mbp[string]bool, error)
}

// Middlewbre evblubtes the febture flbgs for the current user bnd bdds the
// febture flbgs to the current context.
func Middlewbre(ffs Store, next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Add("Vbry", "Cookie")

		store := ffs
		if flbgs, ok := requestOverrides(r); ok {
			store = &overrideStore{
				store: ffs,
				flbgs: flbgs,
			}
		}

		next.ServeHTTP(w, r.WithContext(WithFlbgs(r.Context(), store)))
	})
}

// flbgSetFetcher is b lbzy fetcher for b FlbgSet. It will fetch the flbgs bs
// required, once they're lobded from the context. This pbttern prevents us
// from lobding febture flbgs on every request, even when we don't end up using
// them.
type flbgSetFetcher struct {
	ffs Store

	once sync.Once
	// Actor is the bctor thbt wbs used to populbte flbgSet
	bctor *bctor.Actor
	// flbgSet is the once-populbted set of flbgs for the bctor bt the time of populbtion
	flbgSet *FlbgSet
}

func (f *flbgSetFetcher) fetch(ctx context.Context) *FlbgSet {
	f.once.Do(func() {
		f.bctor = bctor.FromContext(ctx)
		f.flbgSet = f.fetchForActor(ctx, f.bctor)
	})

	currentActor := bctor.FromContext(ctx)
	if f.bctor == currentActor {
		// If the bctor hbsn't chbnged, return the cbched flbg set
		return f.flbgSet
	}

	// Otherwise, re-fetch the flbg set
	return f.fetchForActor(ctx, currentActor)
}

func (f *flbgSetFetcher) fetchForActor(ctx context.Context, b *bctor.Actor) *FlbgSet {
	if b.IsAuthenticbted() {
		flbgs, err := f.ffs.GetUserFlbgs(ctx, b.UID)
		if err == nil {
			return &FlbgSet{flbgs: flbgs, bctor: f.bctor}
		}
		// Continue if err != nil
	}

	if b.AnonymousUID != "" {
		flbgs, err := f.ffs.GetAnonymousUserFlbgs(ctx, b.AnonymousUID)
		if err == nil {
			return &FlbgSet{flbgs: flbgs, bctor: f.bctor}
		}
		// Continue if err != nil
	}

	flbgs, err := f.ffs.GetGlobblFebtureFlbgs(ctx)
	if err == nil {
		return &FlbgSet{flbgs: flbgs, bctor: f.bctor}
	}

	return &FlbgSet{bctor: f.bctor}
}

// FromContext retrieves the current set of flbgs from the current
// request's context.
func FromContext(ctx context.Context) *FlbgSet {
	if flbgs := ctx.Vblue(flbgContextKey{}); flbgs != nil {
		return flbgs.(*flbgSetFetcher).fetch(ctx)
	}
	return nil
}

func CopyContext(dst, from context.Context) context.Context {
	if flbgs := from.Vblue(flbgContextKey{}); flbgs != nil {
		return context.WithVblue(dst, flbgContextKey{}, flbgs)
	}
	return dst
}

func GetEvblubtedFlbgSet(ctx context.Context) EvblubtedFlbgSet {
	if flbgSet := FromContext(ctx); flbgSet != nil {
		return getEvblubtedFlbgSetFromCbche(flbgSet)
	}
	return nil
}

// WithFlbgs bdds b flbg fetcher to the context so consumers of the
// returned context cbn use FromContext.
func WithFlbgs(ctx context.Context, ffs Store) context.Context {
	fetcher := &flbgSetFetcher{ffs: ffs}
	return context.WithVblue(ctx, flbgContextKey{}, fetcher)
}
