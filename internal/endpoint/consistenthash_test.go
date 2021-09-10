package endpoint

import (
	"fmt"
	"hash/crc32"
	"os"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/montanaflynn/stats"
	"github.com/segmentio/fasthash/fnv1"
	"github.com/segmentio/fasthash/fnv1a"
	"github.com/sourcegraph/go-rendezvous"

	"github.com/cespare/xxhash"
)

func TestConsistentHashing(t *testing.T) {
	keys := readKeys(t, "keys.csv")

	type (
		hasher   func(string) string
		hashFunc struct {
			name string
			hash func(nodes []string) hasher
		}
	)

	fns := []*hashFunc{
		{
			name: "consistent(crc32ieee)",
			hash: func(nodes []string) hasher {
				m := hashMapNew(50, crc32.ChecksumIEEE)
				m.add(nodes...)
				return m.Lookup
			},
		},
		{
			name: "rendezvous(xxhash64)",
			hash: func(nodes []string) hasher {
				return rendezvous.New(nodes, xxhash.Sum64String).Lookup
			},
		},
		{
			name: "rendezvous(fnv)",
			hash: func(nodes []string) hasher {
				return rendezvous.New(nodes, fnv1.HashString64).Lookup
			},
		},
		{
			name: "rendezvous(fnva)",
			hash: func(nodes []string) hasher {
				return rendezvous.New(nodes, fnv1a.HashString64).Lookup
			},
		},
	}

	t.Run("Distribution", func(t *testing.T) {
		nodes := makeNodes(24)
		counts := map[string]map[string]int{}

		for _, f := range fns {
			hash := f.hash(nodes)
			for _, k := range keys {
				if counts[f.name] == nil {
					counts[f.name] = make(map[string]int, len(keys))
				}
				counts[f.name][hash(k)]++
			}
		}

		var b strings.Builder
		w := tabwriter.NewWriter(&b, 2, 2, 2, ' ', tabwriter.AlignRight)
		fmt.Fprintln(w)

		for i, node := range nodes {
			fmt.Fprintf(w, "node %2d", i)
			for _, f := range fns {
				fmt.Fprintf(w, "\t%s\t=>\t%d\t\t%.2f%%",
					f.name,
					counts[f.name][node],
					(float64(counts[f.name][node])/float64(len(keys)))*100.0,
				)
			}
			fmt.Fprintln(w)
		}

		fmt.Fprintf(w, "\nStats:\n")
		for _, f := range fns {
			data := make(stats.Float64Data, 0, len(counts[f.name]))
			for _, count := range counts[f.name] {
				data = append(data, float64(count))
			}

			min, _ := data.Min()
			max, _ := data.Max()
			mean, _ := data.Mean()
			median, _ := data.Median()
			stdev, _ := data.StandardDeviation()

			fmt.Fprintf(w, "%s\tmin: %d\tmax: %d\tmean: %.3f\tmedian: %d\t\tstdev: %.3f\n",
				f.name,
				int(min),
				int(max),
				mean,
				int(median),
				stdev,
			)

		}

		w.Flush()
		t.Log(b.String())
	})

	t.Run("ClusterDelta", func(t *testing.T) {
		for _, tc := range [][2]int{
			{24, 25},
			{24, 48},
		} {
			tc := tc
			t.Run(fmt.Sprintf("%d->%d", tc[0], tc[1]), func(t *testing.T) {
				beforeNodes := makeNodes(tc[0])
				afterNodes := makeNodes(tc[1])

				type nodeStats struct {
					before int
					after  int
					in     int
					out    int
					same   int
				}

				stats := map[string]map[string]*nodeStats{}

				for _, f := range fns {
					beforeHash := f.hash(beforeNodes)
					afterHash := f.hash(afterNodes)

					stats[f.name] = map[string]*nodeStats{}

					for _, k := range keys {
						beforeNode := beforeHash(k)
						afterNode := afterHash(k)

						if stats[f.name][beforeNode] == nil {
							stats[f.name][beforeNode] = &nodeStats{}
						}

						if stats[f.name][afterNode] == nil {
							stats[f.name][afterNode] = &nodeStats{}
						}

						stats[f.name][beforeNode].before++
						stats[f.name][afterNode].after++

						if beforeNode != afterNode {
							stats[f.name][beforeNode].out++
							stats[f.name][afterNode].in++
						} else {
							stats[f.name][beforeNode].same++
						}
					}
				}

				var b strings.Builder
				w := tabwriter.NewWriter(&b, 0, 2, 1, ' ', tabwriter.AlignRight)
				fmt.Fprintln(w)

				for i, node := range afterNodes {
					fmt.Fprintf(w, "node %2d", i)
					for _, f := range fns {
						fmt.Fprintf(w, "\t\t%s\t=>\tout: \t%d \t%.2f%%\t in: %d",
							f.name,
							stats[f.name][node].out,
							(float64(stats[f.name][node].out)/float64(stats[f.name][node].before))*100,
							stats[f.name][node].in,
						)
					}
					fmt.Fprintln(w)
				}

				w.Flush()
				t.Log(b.String())
			})
		}
	})

	t.Run("HashFnChange", func(t *testing.T) {
		nodes := makeNodes(24)

		in := map[string]int{}
		out := map[string]int{}
		same := map[string]int{}

		before := fns[0]
		after := fns[1]
		beforeHash := before.hash(nodes)
		afterHash := after.hash(nodes)

		for _, k := range keys {
			beforeNode := beforeHash(k)
			afterNode := afterHash(k)

			if beforeNode != afterNode {
				out[beforeNode]++
				in[afterNode]++
			} else {
				same[beforeNode]++
			}
		}

		var b strings.Builder
		w := tabwriter.NewWriter(&b, 0, 2, 1, ' ', tabwriter.AlignRight)
		fmt.Fprintln(w)

		for i, node := range nodes {
			fmt.Fprintf(w, "node %2d\t%s -> %s\t=>\t-%d\t+%d\t=\t%d\n",
				i,
				before.name,
				after.name,
				out[node],
				in[node],
				same[node],
			)
		}

		w.Flush()
		t.Log(b.String())
	})
}

func BenchmarkConsistenHashing(b *testing.B) {
	for _, nodeCount := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("NodeCount-%d", nodeCount), func(b *testing.B) {
			b.ReportAllocs()
			var nodes []string
			for i := 0; i < nodeCount; i++ {
				nodes = append(nodes, fmt.Sprintf("node%d", i))
			}
			h := newConsistentHash(nodes)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				h.Lookup("foo")
			}
		})
	}
}

func makeNodes(n int) []string {
	nodes := make([]string, n)
	for i := 0; i < n; i++ {
		nodes[i] = fmt.Sprintf("indexed-search-%d.indexed-search:6070", i)
	}
	return nodes
}

func readKeys(t testing.TB, name string) (keys []string) {
	t.Helper()

	data, err := os.ReadFile(name)
	if err != nil {
		t.Skipf("failed to read %s, skipping test: %s", name, err)
		return
	}

	return strings.Split(string(data), "\n")
}
