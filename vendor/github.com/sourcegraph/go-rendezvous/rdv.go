package rendezvous

import "math"

type Rendezvous struct {
	nodes map[string]int
	nstr  []string
	nhash []uint64
	hash  Hasher
}

type Hasher func(s string) uint64

func New(nodes []string, hash Hasher) *Rendezvous {
	r := &Rendezvous{
		nodes: make(map[string]int, len(nodes)),
		nstr:  make([]string, len(nodes)),
		nhash: make([]uint64, len(nodes)),
		hash:  hash,
	}

	for i, n := range nodes {
		r.nodes[n] = i
		r.nstr[i] = n
		r.nhash[i] = hash(n)
	}

	return r
}

func (r *Rendezvous) Lookup(k string) string {
	// short-circuit if we're empty
	if len(r.nodes) == 0 {
		return ""
	}

	khash := r.hash(k)
	midx := 0
	mhash := xorshiftMult64(khash ^ r.nhash[0])

	for i, nhash := range r.nhash[1:] {
		if h := xorshiftMult64(khash ^ nhash); h > mhash {
			midx, mhash = i+1, h
		}
	}

	return r.nstr[midx]
}

func (r *Rendezvous) LookupN(k string, n int) []string {
	if len(r.nodes) == 0 || n <= 0 || n > len(r.nodes) {
		return nil
	}

	khash := r.hash(k)
	hashes := make([]uint64, len(r.nodes))

	for i, nhash := range r.nhash {
		hashes[i] = xorshiftMult64(khash ^ nhash)
	}

	nodes := make([]string, n)
	mhash := uint64(math.MaxUint64) // max hash

	for i := 0; i < n; i++ {
		midx := 0
		for j, h := range hashes {
			if h > hashes[midx] && h < mhash {
				midx = j
			}
		}
		nodes[i], mhash = r.nstr[midx], hashes[midx]
	}

	return nodes
}

func (r *Rendezvous) Nodes() []string {
	return r.nstr
}

func (r *Rendezvous) Add(node string) {
	r.nodes[node] = len(r.nstr)
	r.nstr = append(r.nstr, node)
	r.nhash = append(r.nhash, r.hash(node))
}

func (r *Rendezvous) Remove(node string) {
	// find index of node to remove
	nidx := r.nodes[node]

	// remove from the slices
	l := len(r.nstr)
	r.nstr[nidx] = r.nstr[l]
	r.nstr = r.nstr[:l]

	r.nhash[nidx] = r.nhash[l]
	r.nhash = r.nhash[:l]

	// update the map
	delete(r.nodes, node)
	moved := r.nstr[nidx]
	r.nodes[moved] = nidx
}

func xorshiftMult64(x uint64) uint64 {
	x ^= x >> 12 // a
	x ^= x << 25 // b
	x ^= x >> 27 // c
	return x * 2685821657736338717
}
