package staticcheck

import (
	"testing"

	"honnef.co/go/tools/lint/lintutil"
	"honnef.co/go/tools/lint/testutil"
)

func TestAll(t *testing.T) {
	c := NewChecker()
	testutil.TestAll(t, c, "")
}

func BenchmarkStdlib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := NewChecker()
		_, _, err := lintutil.Lint(c, []string{"std"}, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNetHttp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := NewChecker()
		_, _, err := lintutil.Lint(c, []string{"net/http"}, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
