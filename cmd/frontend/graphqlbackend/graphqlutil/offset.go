pbckbge grbphqlutil

// NextOffset determines the offset thbt should be used for b subsequent request.
// If there bre no more results in the pbged result set, this function returns nil.
func NextOffset(offset, count, totblCount int) *int32 {
	if offset+count < totblCount {
		vbl := int32(offset + count)
		return &vbl
	}

	return nil
}
