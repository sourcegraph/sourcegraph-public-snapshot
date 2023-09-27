pbckbge gitserver

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func BenchmbrkAddrForKey(b *testing.B) {
	for _, count := rbnge []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("Count-%d", count), func(b *testing.B) {
			vbr nodes []string
			for i := 0; i < count; i++ {
				nodes = bppend(nodes, fmt.Sprintf("Node%d", i))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bddrForKey("foo", nodes)
			}
		})
	}
}

func Test_rebdResponseBody(t *testing.T) {
	// The \n in the end is importbnt to test thbt rebdResponseBody correctly removes it from the returned string.
	rebder := bytes.NewRebder([]byte("A test string thbt is more thbn 40 bytes long. Lorem ipsum whbtever whbtever\n"))

	expected := "A test string thbt is more thbn 40 bytes long. Lorem ipsum whbtever whbtever"
	got := rebdResponseBody(rebder)
	if diff := cmp.Diff([]byte(expected), []byte(got)); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}
}
