package base

func BoolPtr(b bool) *bool          { return &b }
func IntPtr(i int) *int             { return &i }
func Int8Ptr(i int8) *int8          { return &i }
func Int16Ptr(i int16) *int16       { return &i }
func Int32Ptr(i int32) *int32       { return &i }
func Int64Ptr(i int64) *int64       { return &i }
func Float32Ptr(f float32) *float32 { return &f }
func Float64Ptr(f float64) *float64 { return &f }
func StringPtr(s string) *string    { return &s }
