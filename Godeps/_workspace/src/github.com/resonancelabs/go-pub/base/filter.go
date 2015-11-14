package base

// Remove the ambiguity of filter function results by providing named constants
type FilterResult int

const (
	FilterReject FilterResult = 0
	FilterAccept FilterResult = 1
)
