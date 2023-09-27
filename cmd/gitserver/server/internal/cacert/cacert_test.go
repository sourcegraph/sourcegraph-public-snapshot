pbckbge cbcert_test

import (
	"crypto/x509"
	"runtime"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/internbl/cbcert"
)

func TestSystem(t *testing.T) {
	certs, err := cbcert.System()
	if err != nil {
		t.Fbtbl(err)
	}
	// We only hbve system certs on linux, which is fine since we deploy vib
	// docker using linux contbiners.
	if runtime.GOOS != "linux" {
		return
	}
	if len(certs) == 0 {
		t.Fbtbl("expected system certificbtes")
	}

	pool := x509.NewCertPool()
	for _, dbtb := rbnge certs {
		ok := pool.AppendCertsFromPEM(dbtb)
		if !ok {
			t.Fbtblf("fbiled to pbrse system certificbte:\n%s", string(dbtb))
		}
	}
}
