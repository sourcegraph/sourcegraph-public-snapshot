package awskms

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/encryption/envelope"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

// testKeyID is the ID of a key defined here:
// https://us-west-2.console.aws.amazon.com/kms/home?region=us-west-2#/kms/keys/4b739277-5a93-4551-b71c-99608c9c805d
// If you want to update this test, feel free to replace the key ID used here.
const testKeyID = "4b739277-5a93-4551-b71c-99608c9c805d"

func TestRoundtrip(t *testing.T) {
	ctx := context.Background()
	// Validate that we successfully worked around the 4096 bytes restriction.
	testString := strings.Repeat("test1234", 4096)
	keyConfig := schema.AWSKMSEncryptionKey{
		KeyId:  testKeyID,
		Region: "us-west-2",
		Type:   "awskms",
	}

	cf, save := newClientFactory(t, "awskms")
	defer save(t)

	// Create http cli with aws defaults.
	cli, err := cf.Doer(func(c *http.Client) error {
		c.Transport = awshttp.NewBuildableClient().GetTransport()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Build config options for aws config.
	configOpts := awsConfigOptsForKeyConfig(keyConfig)
	configOpts = append(configOpts, config.WithHTTPClient(cli))
	configOpts = append(configOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
		readEnvFallback("AWS_ACCESS_KEY_ID", "test"),
		readEnvFallback("AWS_SECRET_ACCESS_KEY", "test"),
		readEnvFallback("AWS_SESION_TOKEN", ""),
	)))
	defaultConfig, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		t.Fatal(err)
	}

	k, err := newKey(ctx, keyConfig, defaultConfig)
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

	// Now test that the old encrypted values still work:
	ciphertext, err := oldEncrypt(ctx, k.client, &k.keyID, []byte(testString))
	require.NoError(t, err)
	res, err = k.Decrypt(ctx, ciphertext)
	require.NoError(t, err)
	if res.Secret() != testString {
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

var shouldUpdate = flag.Bool("update", false, "Update testdata")

func newClientFactory(t testing.TB, name string, mws ...httpcli.Middleware) (*httpcli.Factory, func(testing.TB)) {
	t.Helper()
	cassete := filepath.Join("testdata", strings.ReplaceAll(name, " ", "-"))
	rec := newRecorder(t, cassete, *shouldUpdate)
	mw := httpcli.NewMiddleware(mws...)
	return httpcli.NewFactory(mw, httptestutil.NewRecorderOpt(rec)),
		func(t testing.TB) { save(t, rec) }
}

func newRecorder(t testing.TB, file string, record bool) *recorder.Recorder {
	t.Helper()
	rec, err := httptestutil.NewRecorder(file, record, func(i *cassette.Interaction) error {
		i.Request.Headers.Del("Amz-Sdk-Invocation-Id")
		i.Request.Headers.Del("X-Amz-Date")
		return nil
	})
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

func readEnvFallback(key, fallback string) string {
	if val := os.Getenv(key); val == "" {
		return fallback
	} else {
		return val
	}
}

func oldEncrypt(ctx context.Context, client *kms.Client, keyID *string, plaintext []byte) ([]byte, error) {
	// Encrypt plaintext.
	res, err := client.GenerateDataKey(ctx, &kms.GenerateDataKeyInput{
		KeyId:   keyID,
		KeySpec: types.DataKeySpecAes256,
	})
	if err != nil {
		return nil, err
	}

	ev := encryptedValue{
		Key: res.CiphertextBlob,
	}
	ev.Ciphertext, ev.Nonce, err = aesEncrypt(plaintext, res.Plaintext)
	if err != nil {
		return nil, err
	}

	jsonKey, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}
	buf := base64.StdEncoding.EncodeToString(jsonKey)
	return []byte(buf), err
}

func aesEncrypt(plaintext, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}
