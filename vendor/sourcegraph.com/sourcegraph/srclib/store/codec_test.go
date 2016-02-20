package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"testing"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestCodec(t *testing.T) {
	ns := []int{0, 1, 2, 3, 4, 5, 10, 100, 1000, 5000}
	lengths := map[codec]map[int]int{}
	tests := []struct {
		codec codec
	}{
		{
			codec: ProtobufCodec{},
		},
		{
			codec: JSONCodec{},
		},
	}
	for _, test := range tests {
		for _, n := range ns {
			orig := makeGraphData(t, n)

			var buf bytes.Buffer
			if _, err := test.codec.NewEncoder(&buf).Encode(&orig); err != nil {
				t.Errorf("%T (%d): Encode: %s", test.codec, n, err)
				continue
			}

			if _, present := lengths[test.codec]; !present {
				lengths[test.codec] = map[int]int{}
			}
			lengths[test.codec][n] = buf.Len()
			t.Logf("%T (%d): encoded byte length: %d", test.codec, n, buf.Len())

			var decoded graph.Output
			if _, err := test.codec.NewDecoder(&buf).Decode(&decoded); err != nil {
				t.Errorf("%T (%d): Decode: %s", test.codec, n, err)
				continue
			}

			if !reflect.DeepEqual(orig, decoded) {
				t.Errorf("%T (%d): decoded did not match original", test.codec, n)
				if n < 2 {
					t.Logf("orig (%T)\n%v\n\ndecoded (%T)\n%v", orig, orig, decoded, decoded)
				}
				continue
			}
		}
	}

	for i, test := range tests {
		c1 := test.codec
		for _, n := range ns {
			for _, test2 := range tests[i:] {
				c2 := test2.codec
				if c1 == c2 {
					continue
				}
				t.Logf("%T byte len vs %T (%d): %d/%d (%.1f)", c1, c2, n, lengths[c1][n], lengths[c2][n], float64(lengths[c1][n])/float64(lengths[c2][n])*100)
			}
		}
	}
}

func BenchmarkJSONCodec_Encode_1(b *testing.B)     { benchmarkCodec_Encode(b, JSONCodec{}, 1) }
func BenchmarkJSONCodec_Encode_500(b *testing.B)   { benchmarkCodec_Encode(b, JSONCodec{}, 500) }
func BenchmarkJSONCodec_Encode_5000(b *testing.B)  { benchmarkCodec_Encode(b, JSONCodec{}, 5000) }
func BenchmarkJSONCodec_Encode_50000(b *testing.B) { benchmarkCodec_Encode(b, JSONCodec{}, 50000) }

func BenchmarkJSONCodec_Decode_1(b *testing.B)     { benchmarkCodec_Decode(b, JSONCodec{}, 1) }
func BenchmarkJSONCodec_Decode_500(b *testing.B)   { benchmarkCodec_Decode(b, JSONCodec{}, 500) }
func BenchmarkJSONCodec_Decode_5000(b *testing.B)  { benchmarkCodec_Decode(b, JSONCodec{}, 5000) }
func BenchmarkJSONCodec_Decode_50000(b *testing.B) { benchmarkCodec_Decode(b, JSONCodec{}, 50000) }

func BenchmarkProtobufCodec_Encode_1(b *testing.B)    { benchmarkCodec_Encode(b, ProtobufCodec{}, 1) }
func BenchmarkProtobufCodec_Encode_500(b *testing.B)  { benchmarkCodec_Encode(b, ProtobufCodec{}, 500) }
func BenchmarkProtobufCodec_Encode_5000(b *testing.B) { benchmarkCodec_Encode(b, ProtobufCodec{}, 5000) }
func BenchmarkProtobufCodec_Encode_50000(b *testing.B) {
	benchmarkCodec_Encode(b, ProtobufCodec{}, 50000)
}

func BenchmarkProtobufCodec_Decode_1(b *testing.B)    { benchmarkCodec_Decode(b, ProtobufCodec{}, 1) }
func BenchmarkProtobufCodec_Decode_500(b *testing.B)  { benchmarkCodec_Decode(b, ProtobufCodec{}, 500) }
func BenchmarkProtobufCodec_Decode_5000(b *testing.B) { benchmarkCodec_Decode(b, ProtobufCodec{}, 5000) }
func BenchmarkProtobufCodec_Decode_50000(b *testing.B) {
	benchmarkCodec_Decode(b, ProtobufCodec{}, 50000)
}

func makeGraphData(t testing.TB, n int) graph.Output {
	data := graph.Output{}
	if n > 0 {
		data.Defs = make([]*graph.Def, n)
		data.Refs = make([]*graph.Ref, n)
	}
	for i := 0; i < n; i++ {
		data.Defs[i] = &graph.Def{
			DefKey:   graph.DefKey{Path: fmt.Sprintf("def-path-%d", i)},
			Name:     fmt.Sprintf("def-name-%d", i),
			Kind:     "mykind",
			DefStart: uint32((i % 53) * 37),
			DefEnd:   uint32((i%53)*37 + (i % 20)),
			File:     fmt.Sprintf("dir%d/subdir%d/subsubdir%d/file-%d.foo", i%5, i%3, i%7, i%5),
			Exported: i%5 == 0,
			Local:    i%3 == 0,
			Data:     []byte(`"` + strings.Repeat("abcd", 50) + `"`),
		}
		data.Refs[i] = &graph.Ref{
			DefPath: fmt.Sprintf("ref-path-%d", i),
			Def:     i%5 == 0,
			Start:   uint32((i % 51) * 39),
			End:     uint32((i%51)*37 + (int(i) % 18)),
			File:    fmt.Sprintf("dir%d/subdir%d/subsubdir%d/file-%d.foo", i%3, i%5, i%7, i%5),
		}
		if i%3 == 0 {
			data.Refs[i].DefUnit = fmt.Sprintf("def-unit-%d", i%17)
			data.Refs[i].DefUnitType = fmt.Sprintf("def-unit-type-%d", i%3)
			if i%7 == 0 {
				data.Refs[i].DefRepo = fmt.Sprintf("def-repo-%d", i%13)
			}
		}
	}
	return data
}

func benchmarkCodec_Encode(b *testing.B, c codec, n int) {
	orig := makeGraphData(b, n)
	enc := c.NewEncoder(ioutil.Discard)

	runtime.GC()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := enc.Encode(&orig); err != nil {
			b.Fatalf("%T (%d): Encode: %s", c, n, err)
		}
	}
}

func benchmarkCodec_Decode(b *testing.B, c codec, n int) {
	orig := makeGraphData(b, n)
	var buf bytes.Buffer
	if _, err := c.NewEncoder(&buf).Encode(&orig); err != nil {
		b.Errorf("%T (%d): Encode: %s", c, n, err)
	}
	bb := buf.Bytes()

	runtime.GC()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decoded graph.Output
		if _, err := c.NewDecoder(bytes.NewReader(bb)).Decode(&decoded); err != nil {
			b.Fatalf("%T (%d): Decode: %s", c, n, err)
		}
	}
}
