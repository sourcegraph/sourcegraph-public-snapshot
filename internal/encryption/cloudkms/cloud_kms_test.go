package cloudkms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	googlehttptransport "google.golang.org/api/transport/http"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/sourcegraph/sourcegraph/internal/encryption/envelope"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const testKeyID = "projects/erik-test-kms/locations/us/keyRings/erik-test/cryptoKeys/test-kms"

func TestRoundtrip(t *testing.T) {
	ctx := context.Background()

	// Validate that we successfully worked around size restrictions.
	testString := strings.Repeat("test1234", 128000)

	cf, save := newClientFactory(t, "cloudkms")
	defer save(t)

	opts := clientOptions("")
	if !*shouldUpdate {
		// To not require having a gcloud credentials file locally, we overwrite
		// the settings with some bogus API key when update is false.
		opts = append(opts, option.WithAPIKey("bogus"))
	}

	// Create http cli.
	cli, err := cf.Doer(func(c *http.Client) (err error) {
		cliOpts := append([]option.ClientOption{
			internaloption.WithDefaultEndpoint("https://cloudkms.googleapis.com"),
			internaloption.WithDefaultMTLSEndpoint("https://cloudkms.mtls.googleapis.com"),
			internaloption.WithDefaultAudience("https://cloudkms.googleapis.com/"),
			internaloption.WithDefaultScopes(kms.DefaultAuthScopes()...),
		}, opts...)
		c.Transport, err = googlehttptransport.NewTransport(ctx, http.DefaultTransport, cliOpts...)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	opts = append(opts, option.WithHTTPClient(cli.(*http.Client)))
	client, err := kms.NewKeyManagementRESTClient(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure the secret is stable so that mocked requests always match on the value
	// to be encrypted.
	envelope.MockGenerateSecret = func() ([]byte, error) {
		return []byte("32byteslongsecret..............."), nil
	}
	t.Cleanup(func() {
		envelope.MockGenerateSecret = nil
	})

	k, err := newKeyWithClient(ctx, testKeyID, client)
	if err != nil {
		t.Fatal(err)
	}

	ct, err := k.Encrypt(ctx, []byte(testString))
	if err != nil {
		t.Fatal(err)
	}

	res, err := k.Decrypt(ctx, ct)
	if err != nil {
		t.Fatal(err)
	}

	if res.Secret() != testString {
		t.Fatalf("expected %s, got %s", testString, res.Secret())
	}

	// Verify that old payloads can still be decrypted correctly.
	shortTestString := "verysecrettestvalue"
	oldCiphertext, err := oldEncrypt(ctx, client, k.name, []byte(shortTestString))
	require.NoError(t, err)
	res, err = k.Decrypt(ctx, oldCiphertext)
	require.NoError(t, err)

	if res.Secret() != shortTestString {
		t.Fatalf("expected %s, got %s", testString, res.Secret())
	}

	// And finally, test that empty plaintext works:
	t.Run("empty plaintext", func(t *testing.T) {
		ct, err = k.Encrypt(ctx, []byte(""))
		if err != nil {
			t.Fatal(err)
		}
		res, err = k.Decrypt(ctx, ct)
		if err != nil {
			t.Fatal(err)
		}
		if res.Secret() != "" {
			t.Fatalf("expected %s, got %s", "", res.Secret())
		}
	})
}

func oldEncrypt(ctx context.Context, client *kms.KeyManagementClient, keyName string, plaintext []byte) (_ []byte, err error) {
	// encrypt plaintext
	res, err := client.Encrypt(ctx, &kmspb.EncryptRequest{ //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45843
		Name:            keyName,
		Plaintext:       plaintext,
		PlaintextCrc32C: wrapperspb.Int64(int64(crc32Sum(plaintext))),
	})
	if err != nil {
		return nil, err
	}
	// check that both the plaintext & ciphertext checksums are valid
	if !res.VerifiedPlaintextCrc32C ||
		res.CiphertextCrc32C.GetValue() != int64(crc32Sum(res.Ciphertext)) {
		return nil, errors.New("invalid checksum, request corrupted in transit")
	}
	ek := encryptedValue{
		KeyName:    res.Name,
		Ciphertext: res.Ciphertext,
		Checksum:   crc32Sum(plaintext),
	}
	jsonKey, err := json.Marshal(ek)
	if err != nil {
		return nil, err
	}
	buf := base64.StdEncoding.EncodeToString(jsonKey)
	return []byte(buf), err
}

type encryptedValue struct {
	KeyName    string
	Ciphertext []byte
	Checksum   uint32
}

var shouldUpdate = flag.Bool("update", false, "Update testdata")

func newClientFactory(t testing.TB, name string, mws ...httpcli.Middleware) (*httpcli.Factory, func(testing.TB)) {
	t.Helper()
	cassete := filepath.Join("testdata", strings.ReplaceAll(name, " ", "-"))
	rec := newRecorder(t, cassete, *shouldUpdate)
	mw := httpcli.NewMiddleware(mws...)
	return httpcli.NewFactory(mw, httptestutil.NewRecorderOpt(rec)), func(t testing.TB) { save(t, rec) }
}

func newRecorder(t testing.TB, file string, record bool) *recorder.Recorder {
	t.Helper()
	rec, err := httptestutil.NewRecorder(file, record, func(i *cassette.Interaction) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

func save(t testing.TB, rec *recorder.Recorder) {
	t.Helper()
	if err := rec.Stop(); err != nil {
		t.Errorf("failed to update test data: %s", err)
	}
}
