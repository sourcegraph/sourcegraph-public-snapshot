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

type hashFn func(data []byte) uint32

type hashMap struct {
	hash     hashFn
	replicas int
	keys     []int // Sorted
	hashMap  map[int]string
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

// Adds some keys to the hash.
func (m *hashMap) add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Gets the closest item in the hash to the provided key that is not in
// exclude.
func (m *hashMap) get(key string, exclude map[string]bool) string {
	if m.isEmpty() {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	// Means we have cycled back to the first replica.
	if idx == len(m.keys) {
		idx = 0
	}

	if exclude == nil {
		return m.hashMap[m.keys[idx]]
	}

	// This will return the same key our binary search would if we excluded
	// all keys in exclude.
	for offset := 0; offset < len(m.keys); offset++ {
		item := m.hashMap[m.keys[(idx+offset)%len(m.keys)]]
		if _, ok := exclude[item]; !ok {
			return item
		}
	}
	return ""
}

// getAll returns a slice of values such that m[keys[i]] == values[i]. It is a
// more efficient implementation than calling get on each key.
func (m *hashMap) getAll(keys ...string) []string {
	next := func(offset, hash int) int {
		k := m.keys[offset:]
		idx := sort.Search(len(k), func(i int) bool { return k[i] >= hash })
		return idx + offset
	}
	// Special case large sets of keys. Better to just do a simple linear
	// search if we have roughly more keys to lookup than keys in the map.
	if len(keys) >= len(m.keys)/2 {
		next = func(offset, hash int) int {
			for ; offset < len(m.keys) && m.keys[offset] < hash; offset++ {
			}
			return offset
		}
	}

	type hashed struct {
		idx  int
		hash int
	}
	hashes := make([]hashed, len(keys))
	for i, key := range keys {
		hashes[i] = hashed{idx: i, hash: int(m.hash([]byte(key)))}
	}
	sort.Slice(hashes, func(i, j int) bool { return hashes[i].hash < hashes[j].hash })

	values := make([]string, len(keys))
	offset := 0
	for _, h := range hashes {
		offset = next(offset, h.hash)
		if offset == len(m.keys) {
			values[h.idx] = m.hashMap[m.keys[0]]
		} else {
			values[h.idx] = m.hashMap[m.keys[offset]]
		}
	}
	return values
}
