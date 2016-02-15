// +build !windows

package phtable

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

var (
	sampleData = map[string]string{
		"one":   "1",
		"two":   "2",
		"three": "3",
		"four":  "4",
		"five":  "5",
		"six":   "6",
		"seven": "7",
	}
)

var (
	words [][]byte
)

func init() {
	f, err := os.Open("/usr/share/dict/words")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		words = append(words, line)
	}
}

func TestCHDBuilder(t *testing.T) {
	b := Builder(0)
	for k, v := range sampleData {
		b.Add([]byte(k), []byte(v))
	}
	c, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	c.StoreKeys = true
	if !reflect.DeepEqual(len(c.keys), 15) {
		t.Errorf("got len(c.keys) == %v, want %v", len(c.keys), 15)
	}
	for k, v := range sampleData {
		vv := c.Get([]byte(k))
		if string(vv) != v {
			t.Errorf("got value == %q, want %q", vv, v)
		}
	}
	if v := c.Get([]byte("monkey")); v != nil {
		t.Errorf("for key 'monkey', got value %q, want nil", v)
	}
}

func TestCHDSerialization(t *testing.T) {
	words := words[:100]

	cb := Builder(0)
	for _, v := range words {
		cb.Add([]byte(v), []byte(v))
	}
	m, err := cb.Build()
	if err != nil {
		t.Fatal(err)
	}
	m.StoreKeys = true
	w := &bytes.Buffer{}
	err = m.Write(w)
	if err != nil {
		t.Fatal(err)
	}

	n, err := Mmap(w.Bytes(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.r, n.r) {
		t.Errorf("got m.r == %v, want %v", m.r, n.r)
	}
	if !reflect.DeepEqual(m.indices, n.indices) {
		t.Errorf("got m.indices == %v, want %v", m.indices, n.indices)
	}
	if !reflect.DeepEqual(m.keys, n.keys) {
		t.Errorf("got m.keys == %v, want %v", m.keys, n.keys)
	}
	if !reflect.DeepEqual(m.values, n.values) {
		t.Errorf("got m.values == %v, want %v", m.values, n.values)
	}
	for _, v := range words {
		vv := n.Get([]byte(v))
		if string(vv) != string(v) {
			t.Errorf("got value == %q, want %q", vv, v)
		}
	}
}

func TestCHDSerialization_empty(t *testing.T) {
	cb := Builder(0)
	m, err := cb.Build()
	if err != nil {
		t.Fatal(err)
	}
	m.StoreKeys = true
	w := &bytes.Buffer{}
	err = m.Write(w)
	if err != nil {
		t.Fatal(err)
	}

	n, err := Mmap(w.Bytes(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.r, n.r) {
		t.Errorf("got r == %v, want %v", m.r, n.r)
	}
	if !reflect.DeepEqual(m.indices, n.indices) {
		t.Errorf("got indices == %v, want %v", m.indices, n.indices)
	}
	if !reflect.DeepEqual(m.keys, n.keys) {
		t.Errorf("got keys == %v, want %v", m.keys, n.keys)
	}
	if !reflect.DeepEqual(m.values, n.values) {
		t.Errorf("got values == %v, want %v", m.values, n.values)
	}
}

func TestCHDSerialization_one(t *testing.T) {
	cb := Builder(0)
	cb.Add([]byte("k"), []byte("v"))
	m, err := cb.Build()
	if err != nil {
		t.Fatal(err)
	}
	m.StoreKeys = true
	w := &bytes.Buffer{}
	err = m.Write(w)
	if err != nil {
		t.Fatal(err)
	}

	n, err := Mmap(w.Bytes(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.r, n.r) {
		t.Errorf("got r == %v, want %v", m.r, n.r)
	}
	if !reflect.DeepEqual(m.indices, n.indices) {
		t.Errorf("got indices == %v, want %v", m.indices, n.indices)
	}
	if !reflect.DeepEqual(m.keys, n.keys) {
		t.Errorf("got keys == %v, want %v", m.keys, n.keys)
	}
	if !reflect.DeepEqual(m.values, n.values) {
		t.Errorf("got values == %v, want %v", m.values, n.values)
	}
}

func TestCHDBuild_two(t *testing.T) {
	cb := Builder(0)
	cb.Add([]byte("p"), []byte("v"))
	cb.Add([]byte("p2"), []byte("v"))
	if _, err := cb.Build(); err != nil {
		t.Fatal(err)
	}
}

const maxTestIter = 7

func TestCHDSerialization_collisionSameLength(t *testing.T) {
	for i := uint8(0); i < maxTestIter; i++ {
		for j := uint8(0); j < maxTestIter-1; j++ {
			if i == j {
				continue
			}
			cb := Builder(2)
			cb.Add([]byte{i, j}, []byte("v"))
			cb.Add([]byte{j, i}, []byte("v"))
			m, err := cb.Build()
			if err != nil {
				t.Fatal(err)
			}
			w := &bytes.Buffer{}
			err = m.Write(w)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestCHDSerialization_collisionDiffLength(t *testing.T) {
	for i := uint8(0); i < maxTestIter; i++ {
		for j := uint8(0); j < maxTestIter; j++ {
			cb := Builder(2)
			cb.Add([]byte{i}, []byte("v"))
			cb.Add([]byte{i, j}, []byte("v"))
			m, err := cb.Build()
			if err != nil {
				t.Fatal(err)
			}
			w := &bytes.Buffer{}
			if err := m.Write(w); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func BenchmarkBuiltinMap(b *testing.B) {
	keys := []string{}
	d := map[string]string{}
	for _, bk := range words {
		k := string(bk)
		d[k] = k
		keys = append(keys, k)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d[keys[i%len(keys)]]
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mph := Builder(len(words))
		for _, w := range words {
			mph.Add(w, w)
		}
	}
}

func BenchmarkCHD(b *testing.B) {
	mph := Builder(len(words))
	for _, w := range words {
		mph.Add(w, w)
	}
	h, _ := mph.Build()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Get(words[i%len(words)])
	}
}
