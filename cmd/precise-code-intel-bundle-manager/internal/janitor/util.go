package janitor

// batchIntSlice returns slices of s (in order) at most batchSize in length.
func batchIntSlice(s []int, batchSize int) [][]int {
	batches := [][]int{}
	for len(s) > batchSize {
		batches = append(batches, s[:batchSize])
		s = s[batchSize:]
	}

	if len(s) > 0 {
		batches = append(batches, s)
	}

	return batches
}
