package gitserver

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func BenchmarkAddrForKey(b *testing.B) {
	for _, count := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("Count-%d", count), func(b *testing.B) {
			var nodes []string
			for i := 0; i < count; i++ {
				nodes = append(nodes, fmt.Sprintf("Node%d", i))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				addrForKey("foo", nodes)
			}
		})
	}
}

func Test_readResponseBody(t *testing.T) {
	// The \n in the end is important to test that readResponseBody correctly removes it from the returned string.
	reader := bytes.NewReader([]byte("A test string that is more than 40 bytes long. Lorem ipsum whatever whatever\n"))

	expected := "A test string that is more than 40 bytes long. Lorem ipsum whatever whatever"
	got := readResponseBody(reader)
	if diff := cmp.Diff([]byte(expected), []byte(got)); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}
