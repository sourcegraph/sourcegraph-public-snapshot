package ace

// source represents source for the parsing process.
type source struct {
	base     *File
	inner    *File
	includes []*File
}

// NewSource creates and returns source.
func NewSource(base, inner *File, includes []*File) *source {
	return &source{
		base:     base,
		inner:    inner,
		includes: includes,
	}
}
