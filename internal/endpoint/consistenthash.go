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
	"os"
	"sort"
	"strconv"

	"github.com/cespare/xxhash/v2"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/go-rendezvous"
)

type consistentHash interface {
	Lookup(string) string
	LookupN(string, int) []string
	Nodes() []string
}

func newConsistentHash(nodes []string) consistentHash {
	if os.Getenv("SRC_ENDPOINTS_CONSISTENT_HASH") != "consistent(crc32ieee)" {
		log15.Info("endpoints: using rendezvous hashing")
		return rendezvous.New(nodes, xxhash.Sum64String)
	}
	// 50 replicas and crc32.ChecksumIEEE are the defaults used by
	// groupcache.
	m := hashMapNew(50, crc32.ChecksumIEEE)
	m.add(nodes...)
	return m
}

type hashFn func(data []byte) uint32

type hashMap struct {
	hash     hashFn
	replicas int
	keys     []int // Sorted
	hashMap  map[int]string
	nodes    []string
}

func hashMapNew(replicas int, fn hashFn) *hashMap {
	m := &hashMap{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
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

// Adds some nodes to the hash.
func (m *hashMap) add(nodes ...string) {
	for _, node := range nodes {
		m.nodes = append(m.nodes, node)
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = node
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

func (m *hashMap) Nodes() []string {
	return m.nodes
}
