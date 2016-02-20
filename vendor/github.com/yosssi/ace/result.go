package ace

// result represents a result of the parsing process.
type result struct {
	base     []element
	inner    []element
	includes map[string][]element
}

// newResult creates and returns a result.
func newResult(base []element, inner []element, includes map[string][]element) *result {
	return &result{
		base:     base,
		inner:    inner,
		includes: includes,
	}
}
