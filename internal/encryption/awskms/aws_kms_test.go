package awskms

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"

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
		"",
	)))
	defaultConfig, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		t.Fatal(err)
	}

	k, err := newKey(ctx, keyConfig, defaultConfig)
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
