// B+-tree
//
// The tree we implement here is a B+-tree based on a paper by Ceylan and
// Mihalcea [1].
//
// B+-trees store all values in the leaves. In our case we store trigrams with
// the goal to quickly retrieve a pointer to the posting list for a given
// trigram. We choose the number of trigrams to store at each leaf based on the
// page size, IE we make sure we are able to load a bucket of ngrams with a
// single disk access.
//
// Here is an example of how our B+-tree looks like for a simple input:
//
// input: "hello world", bucketSize=2, v=2
//
// legend: ()=inner node, []=leaf
//
// (level 1)                        (hel, lo_)
//
// (level 2)          (ell)           (llo)           (o_w, irl, red)
//
// (level 3)      [_wo]  [ell]    [hel]  [llo]    [lo_] [o_w] [orl] [rld, wor]
//
// The leafs are stored as part of the index on disk (mmaped) while all inner
// nodes are loaded into memory when we load the shard.
//
// [1] H. Ceylan and R. Mihalcea. 2011. An Efficient Indexer for Large N-Gram
// Corpora, Proceedings of the ACL-HLT 2011 System Demonstrations, pages
// 103-108

package zoekt

import (
	"encoding/binary"
	"fmt"
	"sort"
)

// btreeBucketSize should be chosen such that in the final tree the buckets are
// as close to the page size as possible, but not above. We insert ngrams in
// order(!), which means after a split of a leaf, the left leaf is not affected
// by further inserts and its size is fixed to bucketSize/2. The rightmost leaf
// might store up to btreeBucketSize ngrams, but the expected size is
// btreeBucketSize/2, too.
//
// On linux "getconf PAGESIZE" returns the number of bytes in a memory page.
const btreeBucketSize = (4096 * 2) / ngramEncoding

type btree struct {
	root node
	opts btreeOpts

	lastBucketIndex int
}

type btreeOpts struct {
	// How many ngrams can be stored at a leaf node.
	bucketSize int
	// all inner nodes, except root, have [v, 2v] children. In the literature,
	// b-trees are inconsistently categorized either by the number of children
	// or by the number of keys. We choose the former.
	v int
}

func newBtree(opts btreeOpts) *btree {
	return &btree{
		root: &leaf{},
		opts: opts,
	}
}

// insert inserts ng into bt.
//
// Note: when all inserts are done, freeze must be called.
func (bt *btree) insert(ng ngram) {
	if leftNode, rightNode, newKey, ok := bt.root.maybeSplit(bt.opts); ok {
		bt.root = &innerNode{keys: []ngram{newKey}, children: []node{leftNode, rightNode}}
	}
	bt.root.insert(ng, bt.opts)
}

// find returns the tuple (bucketIndex, postingIndexOffset), both of which are
// stored at the leaf level. They are effectively pointers to the bucket and
// the posting lists for ngrams stored in the bucket. Since ngrams and their
// posting lists are stored in order, knowing the index of the posting list of
// the first item in the bucket is sufficient.
func (bt *btree) find(ng ngram) (int, int) {
	if bt.root == nil {
		return -1, -1
	}
	return bt.root.find(ng)
}

func (bt *btree) visit(f func(n node)) {
	bt.root.visit(f)
}

// freeze must be called once we are done inserting. It backfills "pointers" to
// the buckets and posting lists.
func (bt *btree) freeze() {
	// Note: Instead of backfilling we could maintain state during insertion,
	// however the visitor pattern seems more natural and shouldn't be a
	// performance issue, because, based on the typical number of trigrams
	// (500k) per shard, the b-trees we construct here only have around 1000
	// leaf nodes.
	offset, bucketIndex := 0, 0
	bt.visit(func(no node) {
		switch n := no.(type) {
		case *leaf:
			n.bucketIndex = bucketIndex
			bucketIndex++

			n.postingIndexOffset = offset
			offset += n.bucketSize
		case *innerNode:
			return
		}
	})

	bt.lastBucketIndex = bucketIndex - 1
}

func (bt *btree) sizeBytes() int {
	sz := 2 * 8 // opts

	sz += int(interfaceBytes)

	bt.visit(func(n node) {
		sz += n.sizeBytes()
	})

	return sz
}

type node interface {
	insert(ng ngram, opts btreeOpts)
	maybeSplit(opts btreeOpts) (left node, right node, newKey ngram, ok bool)
	find(ng ngram) (int, int)
	visit(func(n node))
	sizeBytes() int
}

type innerNode struct {
	keys     []ngram
	children []node
}

type leaf struct {
	bucketIndex int
	// postingIndexOffset is the index of the posting list of the first ngram
	// in the bucket. This is enough to determine the index of the posting list
	// for every other key in the bucket.
	postingIndexOffset int

	// Optimization: Because we insert ngrams in order, we don't actually have
	// to fill the buckets. We just have to keep track of the size of the
	// bucket, so we know when to split, and the key that we have to propagate
	// up to the parent node when we split.
	//
	// If in the future we decide to mutate buckets, we have to replace
	// bucketSize and splitKey by []ngram.
	bucketSize int
	splitKey   ngram
}

func (n *innerNode) sizeBytes() int {
	return len(n.keys)*ngramEncoding + len(n.children)*int(interfaceBytes)
}

func (n *leaf) sizeBytes() int {
	return 4 * 8
}

func (n *leaf) insert(ng ngram, opts btreeOpts) {
	n.bucketSize++

	if n.bucketSize == (opts.bucketSize/2)+1 {
		n.splitKey = ng
	}
}

func (n *innerNode) insert(ng ngram, opts btreeOpts) {
	insertAt := func(i int) {
		// Invariant: Nodes always have a free slot.
		//
		// We split full nodes on the the way down to the leaf. This has the
		// advantage that inserts are handled in a single pass.
		if leftNode, rightNode, newKey, ok := n.children[i].maybeSplit(opts); ok {
			n.keys = append(n.keys[0:i], append([]ngram{newKey}, n.keys[i:]...)...)
			n.children = append(n.children[0:i], append([]node{leftNode, rightNode}, n.children[i+1:]...)...)

			// A split might shift the target index by 1.
			if ng >= n.keys[i] {
				i++
			}
		}
		n.children[i].insert(ng, opts)
	}

	for i, k := range n.keys {
		if ng < k {
			insertAt(i)
			return
		}
	}
	insertAt(len(n.children) - 1)
}

// See btree.find
func (n *innerNode) find(ng ngram) (int, int) {
	for i, k := range n.keys {
		if ng < k {
			return n.children[i].find(ng)
		}
	}
	return n.children[len(n.children)-1].find(ng)
}

// See btree.find
func (n *leaf) find(ng ngram) (int, int) {
	return int(n.bucketIndex), int(n.postingIndexOffset)
}

func (n *leaf) maybeSplit(opts btreeOpts) (left node, right node, newKey ngram, ok bool) {
	if n.bucketSize < opts.bucketSize {
		return
	}
	return &leaf{bucketSize: opts.bucketSize / 2},
		&leaf{bucketSize: opts.bucketSize / 2},
		n.splitKey,
		true
}

func (n *innerNode) maybeSplit(opts btreeOpts) (left node, right node, newKey ngram, ok bool) {
	if len(n.children) < 2*opts.v {
		return
	}
	return &innerNode{
			keys:     append(make([]ngram, 0, opts.v-1), n.keys[0:opts.v-1]...),
			children: append(make([]node, 0, opts.v), n.children[:opts.v]...),
		},
		&innerNode{
			keys:     append(make([]ngram, 0, (2*opts.v)-1), n.keys[opts.v:]...),
			children: append(make([]node, 0, 2*opts.v), n.children[opts.v:]...),
		},
		n.keys[opts.v-1],
		true
}

func (n *leaf) visit(f func(n node)) {
	f(n)
	return
}

func (n *innerNode) visit(f func(n node)) {
	f(n)
	for _, child := range n.children {
		child.visit(f)
	}
}

func (bt *btree) String() string {
	s := ""
	s += fmt.Sprintf("%+v", bt.opts)
	bt.root.visit(func(n node) {
		switch nd := n.(type) {
		case *leaf:
			return
		case *innerNode:
			s += fmt.Sprintf("[")
			for _, key := range nd.keys {
				s += fmt.Sprintf("%d,", key)
			}
			s = s[:len(s)-1] // remove trailing comma
			s += fmt.Sprintf("]")

		}
	})
	return s
}

type btreeIndex struct {
	bt *btree

	// We need the index to read buckets into memory.
	file IndexFile

	// buckets
	ngramSec simpleSection

	postingIndex simpleSection
}

// SizeBytes returns how much memory this structure uses in the heap.
func (b btreeIndex) SizeBytes() (sz int) {
	// btree
	if b.bt != nil {
		sz += int(pointerSize) + b.bt.sizeBytes()
	}
	// ngramSec
	sz += 8
	// postingIndex
	sz += 8
	// postingDataSentinelOffset
	sz += 4
	return
}

// Get returns the simple section of the posting list associated with the
// ngram. The logic is as follows:
// 1. Search the inner nodes to find the bucket that may contain ng (in MEM)
// 2. Read the bucket from disk (1 disk access)
// 3. Binary search the bucket (in MEM)
// 4. Return the simple section pointing to the posting list (in MEM)
func (b btreeIndex) Get(ng ngram) (ss simpleSection) {
	if b.bt == nil {
		return simpleSection{}
	}

	// find bucket
	bucketIndex, postingIndexOffset := b.bt.find(ng)

	// read bucket into memory
	off, sz := b.getBucket(bucketIndex)
	bucket, err := b.file.Read(off, sz)
	if err != nil {
		return simpleSection{}
	}

	// find ngram in bucket
	getNGram := func(i int) ngram {
		i *= ngramEncoding
		return ngram(binary.BigEndian.Uint64(bucket[i : i+ngramEncoding]))
	}

	bucketSize := len(bucket) / ngramEncoding
	x := sort.Search(bucketSize, func(i int) bool {
		return ng <= getNGram(i)
	})

	// return associated posting list
	if x >= bucketSize || getNGram(x) != ng {
		return simpleSection{}
	}

	return b.getPostingList(postingIndexOffset + x)
}

// getPostingList returns the simple section pointing to the posting list of
// the ngram at ngramIndex.
//
// Assumming we don't hit a page boundary, which should be rare given that we
// only read 8 bytes, we need 1 disk access to read the posting offset.
func (b btreeIndex) getPostingList(ngramIndex int) simpleSection {
	relativeOffsetBytes := uint32(ngramIndex) * 4

	if relativeOffsetBytes+8 <= b.postingIndex.sz {
		// read 2 offsets
		o, err := b.file.Read(b.postingIndex.off+relativeOffsetBytes, 8)
		if err != nil {
			return simpleSection{}
		}

		start := binary.BigEndian.Uint32(o[0:4])
		end := binary.BigEndian.Uint32(o[4:8])
		return simpleSection{
			off: start,
			sz:  end - start,
		}
	} else {
		// last ngram => read 1 offset and calculate the size of the posting
		// list from the offset of index section.
		o, err := b.file.Read(b.postingIndex.off+relativeOffsetBytes, 4)
		if err != nil {
			return simpleSection{}
		}

		start := binary.BigEndian.Uint32(o[0:4])
		return simpleSection{
			off: start,
			// The layout of the posting list compound section on disk is
			//
			//                      start       b.postingIndex.off
			//                      v           v
			// [[posting lists (simple section)][index (simple section)]]
			//                      <---------->
			//                    last posting list
			//
			sz: b.postingIndex.off - start,
		}
	}
}

func (b btreeIndex) getBucket(bucketIndex int) (off uint32, sz uint32) {
	// All but the rightmost bucket have exactly bucketSize/2 ngrams
	sz = uint32(b.bt.opts.bucketSize / 2 * ngramEncoding)
	off = b.ngramSec.off + uint32(bucketIndex)*sz

	// Rightmost bucket has size upto the end of the ngramSec.
	if bucketIndex == b.bt.lastBucketIndex {
		sz = b.ngramSec.off + b.ngramSec.sz - off
	}

	return
}

// DumpMap is a debug method which returns the btree as an in-memory
// representation. This is how zoekt represents the ngram index in
// google/zoekt.
func (b btreeIndex) DumpMap() map[ngram]simpleSection {
	if b.bt == nil {
		return nil
	}

	m := make(map[ngram]simpleSection, b.ngramSec.sz/ngramEncoding)

	b.bt.visit(func(no node) {
		switch n := no.(type) {
		case *leaf:
			// read bucket into memory
			off, sz := b.getBucket(n.bucketIndex)
			bucket, _ := b.file.Read(off, sz)

			// decode all ngrams in the bucket and fill map
			for i := 0; i < len(bucket)/ngramEncoding; i++ {
				gram := ngram(binary.BigEndian.Uint64(bucket[i*8:]))
				m[gram] = b.getPostingList(int(n.postingIndexOffset) + i)
			}
		case *innerNode:
			return
		}
	})

	return m
}
