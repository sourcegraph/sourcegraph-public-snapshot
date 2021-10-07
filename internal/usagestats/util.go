package usagestats

func int32Ptr(v int) *int32 {
	v32 := int32(v)
	return &v32
}
