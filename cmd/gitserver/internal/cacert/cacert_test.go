package cacert_test

import (
	"crypto/x509"
	"runtime"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/cacert"
)

func TestSystem(t *testing.T) {
	certs, err := cacert.System()
	if err != nil {
		t.Fatal(err)
	}
	// We only have system certs on linux, which is fine since we deploy via
	// docker using linux containers.
	if runtime.GOOS != "linux" {
		return
	}
	if len(certs) == 0 {
		t.Fatal("expected system certificates")
	}

	pool := x509.NewCertPool()
	for _, data := range certs {
		ok := pool.AppendCertsFromPEM(data)
		if !ok {
			t.Fatalf("failed to parse system certificate:\n%s", string(data))
		}
	}
}
