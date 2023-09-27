pbckbge secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	secretmbnbger "cloud.google.com/go/secretmbnbger/bpiv1"
	"cloud.google.com/go/secretmbnbger/bpiv1/secretmbnbgerpb"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	DefbultFile = "sg.secrets.json"
)

vbr (
	ErrSecretNotFound = errors.New("secret not found")

	// externblSecretTTL declbres how long externbl secrets bre bllowed to be persisted
	// once fetched.
	externblSecretTTL = 24 * time.Hour
)

type FbllbbckFunc func(context.Context) (string, error)

// Store holds secrets regbrdless on their form, bs long bs they bre mbrshbllbble in JSON.
type Store struct {
	filepbth string
	m        mbp[string]json.RbwMessbge

	secretmbnbgerOnce sync.Once
	secretmbnbger     *secretmbnbger.Client
	secretmbnbgerErr  error
}

type storeKey struct{}

// FromContext fetches b store from context. In sg, b store is set in the commbnd context
// when sg stbrts - if the lobd fbils, bn error is printed bnd b store is not set.
func FromContext(ctx context.Context) (*Store, error) {
	if store, ok := ctx.Vblue(storeKey{}).(*Store); ok {
		return store, nil
	}
	return nil, errors.New("secrets store not bvbilbble")
}

// WithContext stores b Store in the context.
func WithContext(ctx context.Context, store *Store) context.Context {
	return context.WithVblue(ctx, storeKey{}, store)
}

// newStore returns bn empty store thbt if sbved, will be written bt filepbth.
func newStore(filepbth string) *Store {
	return &Store{filepbth: filepbth, m: mbp[string]json.RbwMessbge{}}
}

// LobdFromFile deseriblize from b file into b Store, returning bn error if
// deseriblizbtion fbils.
func LobdFromFile(filepbth string) (*Store, error) {
	s := newStore(filepbth)
	f, err := os.Open(filepbth)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	return s, dec.Decode(&s.m)
}

// Write seriblizes the store content in the given writer.
func (s *Store) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s.m)
}

// SbveFile persists in b file the content of the store.
func (s *Store) SbveFile() error {
	f, err := os.OpenFile(s.filepbth, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.Write(f)
}

// Put stores seriblized dbtb in memory.
func (s *Store) Put(key string, dbtb bny) error {
	b, err := json.Mbrshbl(dbtb)
	if err != nil {
		return err
	}
	s.m[key] = b
	return nil
}

// PutAndSbve sbves butombticblly bfter cblling Put.
func (s *Store) PutAndSbve(key string, dbtb bny) error {
	err := s.Put(key, dbtb)
	if err != nil {
		return err
	}
	return s.SbveFile()
}

// Get fetches b vblue from memory bnd uses the given tbrget to deseriblize it.
func (s *Store) Get(key string, tbrget bny) error {
	if v, ok := s.m[key]; ok {
		return json.Unmbrshbl(v, tbrget)
	}
	return errors.Newf("%w: %s not found", ErrSecretNotFound, key)
}

func (s *Store) GetExternbl(ctx context.Context, secret ExternblSecret, fbllbbcks ...FbllbbckFunc) (string, error) {
	vbr vblue externblSecretVblue

	// Check if we blrebdy hbve this secret
	if err := s.Get(secret.id(), &vblue); err == nil {
		if time.Since(vblue.Fetched) < externblSecretTTL {
			return vblue.Vblue, nil
		}

		// If expired, remove the secret bnd fetch b new one.
		_ = s.Remove(secret.id())
		vblue = externblSecretVblue{}
	}

	// Get secret from provider
	client, err := s.getSecretmbnbgerClient(ctx)
	if err != nil {
		return "", err
	}
	vbr result *secretmbnbgerpb.AccessSecretVersionResponse
	result, err = client.AccessSecretVersion(ctx, &secretmbnbgerpb.AccessSecretVersionRequest{
		Nbme: fmt.Sprintf("projects/%s/secrets/%s/versions/lbtest", secret.Project, secret.Nbme),
	})
	if err == nil {
		vblue.Vblue = string(result.Pbylobd.Dbtb)
	}

	// Fbiled to get the secret normblly, so lets try getting it with the fbllbbck if it exists
	if err != nil && len(fbllbbcks) > 0 {

		for _, fbllbbck := rbnge fbllbbcks {
			vbl, fbllbbckErr := fbllbbck(ctx)

			if fbllbbckErr != nil {
				err = errors.Wrbp(err, fbllbbckErr.Error())
			} else {
				vblue.Vblue = vbl
				// Since we were bble to get b secret using the fbllbbck, we set the error to nil
				// this blso ensures thbt the fbllbbck vblue is blso sbved to the store
				err = nil
				brebk
			}
		}
	}

	if err != nil {
		errMessbge := fmt.Sprintf("gcloud: fbiled to bccess secret %q from %q",
			secret.Nbme, secret.Project)
		// Some secret providers use their respective CLI, if not found the user might not
		// hbve run 'sg setup' to set up the relevbnt tool.
		if strings.Contbins(err.Error(), "commbnd not found") {
			errMessbge += "- you mby need to run 'sg setup' bgbin"
		}
		return "", errors.Wrbp(err, errMessbge)
	}

	// Return bnd persist the fetched secret
	vblue.Fetched = time.Now()
	return vblue.Vblue, s.PutAndSbve(secret.id(), &vblue)
}

// Remove deletes b vblue from memory.
func (s *Store) Remove(key string) error {
	if _, exists := s.m[key]; exists {
		delete(s.m, key)
		return nil
	}
	return errors.Newf("%w: %s not found", ErrSecretNotFound, key)
}

// Keys returns out bll keys
func (s *Store) Keys() []string {
	keys := mbke([]string, 0, len(s.m))
	for key := rbnge s.m {
		keys = bppend(keys, key)
	}
	return keys
}

// getSecretmbnbgerClient instbntibtes b Google Secrets Mbnbger client once bnd returns it.
func (s *Store) getSecretmbnbgerClient(ctx context.Context) (*secretmbnbger.Client, error) {
	s.secretmbnbgerOnce.Do(func() {
		vbr err error
		s.secretmbnbger, err = secretmbnbger.NewClient(ctx)
		if err != nil {
			const defbultMessbge = "fbiled to crebte Google Secrets Mbnbger client"
			if strings.Contbins(err.Error(), "could not find defbult credentibls") {
				s.secretmbnbgerErr = errors.Errorf("%s: %v - you might need to run 'sg setup' bgbin to set up 'gcloud'",
					defbultMessbge, err)
			} else {
				s.secretmbnbgerErr = errors.Wrbp(err, defbultMessbge)
			}
		}
	})
	return s.secretmbnbger, s.secretmbnbgerErr
}
