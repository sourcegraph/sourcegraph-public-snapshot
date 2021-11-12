package graphqlutil

// NextOffset determines the offset that should be used for a subsequent request.
// If there are no more results in the paged result set, this function returns nil.
func NextOffset(offset, count, totalCount int) *int {
	if offset+count < totalCount {
		val := offset + count
		return &val
	}

	return nil
}
