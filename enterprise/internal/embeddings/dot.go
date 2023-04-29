package embeddings

func Dot(a, b []int8) int32 {
	if haveArchDot {
		return archDot(a, b)
	}
	return dotPortable(a, b)
}

func dotPortable(a, b []int8) int32 {
	similarity := int32(0)

	count := len(a)
	if count > len(b) {
		// Do this ahead of time so the compiler doesn't need to bounds check
		// every time we index into b.
		panic("mismatched vector lengths")
	}

	i := 0
	for ; i+3 < count; i += 4 {
		m0 := int32(a[i]) * int32(b[i])
		m1 := int32(a[i+1]) * int32(b[i+1])
		m2 := int32(a[i+2]) * int32(b[i+2])
		m3 := int32(a[i+3]) * int32(b[i+3])
		similarity += (m0 + m1 + m2 + m3)
	}

	for ; i < count; i++ {
		similarity += int32(a[i]) * int32(b[i])
	}

	return similarity
}
