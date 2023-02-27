package database

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestEncryptAuthenticator(t *testing.T) {
	ctx := context.Background()

	t.Run("errors", func(t *testing.T) {
		for name, tc := range map[string]struct {
			key encryption.Key
			a   auth.Authenticator
		}{
			"bad authenticator": {
				key: et.TestKey{},
				a:   &badAuthenticator{},
			},
			"bad encrypter": {
				key: &et.BadKey{Err: errors.New("encryption is bad")},
				a:   &auth.BasicAuth{},
			},
		} {
			t.Run(name, func(t *testing.T) {
				if _, _, err := EncryptAuthenticator(ctx, tc.key, tc.a); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		enc := &mockKey{}
		a := &auth.BasicAuth{
			Username: "foo",
			Password: "bar",
		}

		want, err := json.Marshal(struct {
			Type AuthenticatorType
			Auth auth.Authenticator
		}{
			Type: AuthenticatorTypeBasicAuth,
			Auth: a,
		})
		if err != nil {
			t.Fatal(err)
		}

		if have, _, err := EncryptAuthenticator(ctx, enc, a); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if diff := cmp.Diff(string(have), string(want)); diff != "" {
			t.Errorf("unexpected byte slice (-have +want):\n%s", diff)
		}

		if enc.called != 1 {
			t.Errorf("mock encrypter called an unexpected number of times: have=%d want=1", enc.called)
		}
	})
}

type mockKey struct {
	called int
}

var _ encryption.Key = &mockKey{}

func (me *mockKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{}, nil
}

func (me *mockKey) Encrypt(ctx context.Context, value []byte) ([]byte, error) {
	me.called++
	return value, nil
}

func (me *mockKey) Decrypt(ctx context.Context, value []byte) (*encryption.Secret, error) {
	return nil, nil
}

type badAuthenticator struct{}

var _ auth.Authenticator = &badAuthenticator{}

func (*badAuthenticator) Authenticate(*http.Request) error {
	return errors.New("never called")
}

func (*badAuthenticator) Hash() string {
	return "never called"
}
