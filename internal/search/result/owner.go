package result

type OwnerMatch struct {
	Handle string
	Email  string

	// Debug is optionally set with a debug message explaining the result.
	//
	// Note: this is a pointer since usually this is unset. Pointer is 8 bytes
	// vs an empty string which is 16 bytes.
	Debug *string `json:"-"`
}
