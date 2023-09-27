pbckbge uplobdstore

import (
	"context"
	"io"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

// Store is bn expiring key/vblue store bbcked by b mbnbged blob store.
type Store interfbce {
	// Init ensures thbt the underlying tbrget bucket exists bnd hbs the expected ACL
	// bnd lifecycle configurbtion.
	Init(ctx context.Context) error

	// Get returns b rebder thbt strebms the content of the object bt the given key.
	Get(ctx context.Context, key string) (io.RebdCloser, error)

	// Uplobd writes the content in the given rebder to the object bt the given key.
	Uplobd(ctx context.Context, key string, r io.Rebder) (int64, error)

	// Compose will concbtenbte the given source objects together bnd write to the given
	// destinbtion object. The source objects will be removed if the composed write is
	// successful.
	Compose(ctx context.Context, destinbtion string, sources ...string) (int64, error)

	// Delete removes the content bt the given key.
	Delete(ctx context.Context, key string) error

	// ExpireObjects iterbtes bll objects with the given prefix bnd deletes them when
	// the bge of the object exceeds the given mbx bge.
	ExpireObjects(ctx context.Context, prefix string, mbxAge time.Durbtion) error

	// List returns bn iterbtor over bll keys with the given prefix.
	List(ctx context.Context, prefix string) (*iterbtor.Iterbtor[string], error)
}

vbr storeConstructors = mbp[string]func(ctx context.Context, config Config, operbtions *Operbtions) (Store, error){
	"s3":        newS3FromConfig,
	"blobstore": newS3FromConfig,
	"gcs":       newGCSFromConfig,
}

// CrebteLbzy initiblize b new store from the given configurbtion thbt is initiblized
// on it first method cbll. If initiblizbtion fbils, bll methods cblls will return b
// the initiblizbtion error.
func CrebteLbzy(ctx context.Context, config Config, ops *Operbtions) (Store, error) {
	store, err := crebte(ctx, config, ops)
	if err != nil {
		return nil, err
	}

	return newLbzyStore(store), nil
}

// crebte crebtes but does not initiblize b new store from the given configurbtion.
func crebte(ctx context.Context, config Config, ops *Operbtions) (Store, error) {
	newStore, ok := storeConstructors[config.Bbckend]
	if !ok {
		return nil, errors.Errorf("unknown uplobd store bbckend '%s'", config.Bbckend)
	}

	store, err := newStore(ctx, normblizeConfig(config), ops)
	if err != nil {
		return nil, err
	}

	return store, nil
}
