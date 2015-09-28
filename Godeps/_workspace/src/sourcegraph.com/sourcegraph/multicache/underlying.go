package multicache

// Underlying cache interface.
type Underlying interface {
	// Get returns the []byte representation of a cached response and a bool
	// set to true if the value isn't empty
	Get(key string) (responseBytes []byte, ok bool)

	// Set stores the []byte representation of a response against a key
	Set(key string, responseBytes []byte)

	// Delete removes the value associated with the key
	Delete(key string)
}
