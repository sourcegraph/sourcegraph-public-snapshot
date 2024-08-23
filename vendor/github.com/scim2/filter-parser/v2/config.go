package filter

// config represents the internal config of the parser functions.
type config struct {
	// useNumber indicates that json.Number needs to be returned instead of int/float64 values.
	useNumber bool
}
