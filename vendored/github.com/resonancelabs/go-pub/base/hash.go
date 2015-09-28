package base

const (
	// Constants and routines stolen from Mumurmurhash2, 64-bit version
	// https://code.google.com/p/smhasher/source/browse/trunk/MurmurHash2.cpp
	m = 0xc6a4a7935bd1e995
	r = 47
)

/// Bit mixes a 64-bit integer.
func BitMix64(v uint64) int64 {
	h := uint64(v)
	h ^= h >> r
	h *= m
	h ^= h >> r
	return int64(h)
}

/// Hashes two 64-bit integers together.  Returns positive integers only.
func HashInts64(v, seed int64) int64 {
	h := uint64(seed) ^ m
	k := uint64(v)
	k *= m
	k ^= k >> r
	k *= m
	h ^= k
	h *= m

	h ^= h >> r
	h *= m
	h ^= h >> r

	// Mask off the sign bit.
	return int64(h & 0x7fffffffffffffff)
}

func HashInts(v, seed int) int {
	// How do I know what size int is?!
	return int(HashInts64(int64(v), int64(seed)))
}
