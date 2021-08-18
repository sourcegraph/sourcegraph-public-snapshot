/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package endpoint

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type consistentHash interface {
	Lookup(string) string
	LookupN(string, int) []string
	Nodes() map[string]struct{}
}

type hashFn func(data []byte) uint32

type hashMap struct {
	hash     hashFn
	replicas int
	keys     []int // Sorted
	hashMap  map[int]string
	values   map[string]struct{}
}

func hashMapNew(replicas int, fn hashFn) *hashMap {
	m := &hashMap{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
		values:   make(map[string]struct{}),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Returns true if there are no items available.
func (m *hashMap) isEmpty() bool {
	return len(m.keys) == 0
}

// Adds some keys to the hash.
func (m *hashMap) add(keys ...string) {
	for _, key := range keys {
		m.values[key] = struct{}{}
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *hashMap) Lookup(key string) string {
	if m.isEmpty() {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	return m.hashMap[m.keys[idx%len(m.keys)]]
}

func (m *hashMap) LookupN(key string, n int) []string {
	if m.isEmpty() {
		return nil
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	nodes := make([]string, 0, n)
	for offset := 0; offset < n; offset++ {
		nodes = append(nodes, m.hashMap[m.keys[(idx+offset)%len(m.keys)]])
	}

	return nodes
}

func (m *hashMap) Nodes() map[string]struct{} {
	return m.values
}
