package pipeline

// Map is a Pipeline that allows modifications of individual lines from streamline.Stream.
// Implementations can return a nil []byte to indicate a line is to be skipped.
type Map func(line []byte) []byte

var _ Pipeline = (Map)(nil)

func (m Map) ProcessLine(line []byte) ([]byte, error) {
	return m(line), nil
}

// MapErr is a Pipeline that allows modifications of individual lines from
// streamline.Stream with error handling. Implementations can return a nil []byte to
// indicate a line is to be skipped.
//
// Errors interrupt line processing and are propagated to streamline.Stream.
type MapErr func(line []byte) ([]byte, error)

var _ Pipeline = (MapErr)(nil)

func (m MapErr) ProcessLine(line []byte) ([]byte, error) {
	return m(line)
}

// MapIdx is a Pipeline that allows modifications of individual lines from
// streamline.Stream based on the index of each line (i.e. how many lines the Pipeline has
// processed). The first line to be processed has an index of 0.
func MapIdx(mapper func(i int, line []byte) ([]byte, error)) Pipeline {
	return &idxMapper{mapper: mapper}
}

type idxMapper struct {
	mapper func(i int, line []byte) ([]byte, error)
	index  int
}

func (i *idxMapper) ProcessLine(line []byte) ([]byte, error) {
	i.index += 1
	return i.mapper(i.index-1, line)
}
