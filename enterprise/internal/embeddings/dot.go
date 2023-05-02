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

func DotFloat32(row []float32, query []float32) float32 {
	similarity := float32(0)

	count := len(row)
	if count > len(query) {
		// Do this ahead of time so the compiler doesn't need to bounds check
		// every time we index into query.
		panic("mismatched vector lengths")
	}

	i := 0
	for ; i+3 < count; i += 4 {
		m0 := row[i] * query[i]
		m1 := row[i+1] * query[i+1]
		m2 := row[i+2] * query[i+2]
		m3 := row[i+3] * query[i+3]
		similarity += (m0 + m1 + m2 + m3)
	}

	for ; i < count; i++ {
		similarity += row[i] * query[i]
	}

	return similarity
}
