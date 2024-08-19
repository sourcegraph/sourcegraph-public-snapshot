package pipeline

// Pipeline implementations are used to transform the data provided to a streamline.Stream.
// For example, they are useful for mapping and pruning data. To configure a Stream to use
// a Pipeline, use (*Stream).WithPipeline(...).
//
// Note that generally a Pipeline should not be used to implement handling of data - use
// (*Stream).Stream(...) and (*Stream).StreamBytes(...) instead.
type Pipeline interface {
	// ProcessLine returns a modified, unmodified, or omitted line. To omit a line, return
	// a nil []byte - an empty []byte will cause an empty line to be retained.
	//
	// Implementations must not retain line.
	ProcessLine(line []byte) ([]byte, error)
}

// MultiPipeline is a Pipeline that applies all its Pipelines in serial.
type MultiPipeline []Pipeline

var _ Pipeline = (MultiPipeline)(nil)

// ProcessLine will provide the line to all active pipelines in the MultiPipeline in
// serial, passing the result of each pipeline to the next. If any pipeline indicates a
// line should be skipped by returning a nil line, then ProcessLine returns immediately.
func (mp MultiPipeline) ProcessLine(line []byte) ([]byte, error) {
	var err error
	for _, p := range mp {
		line, err = p.ProcessLine(line)
		if err != nil {
			break
		}
		// If the line returned is nil, we have nothing to pass on to next pipeline, since
		// nil indicates we should skip the line entirely. A zero-length line is still
		// valid and should still be provided to the next processor.
		if line == nil {
			break
		}
	}
	return line, err
}
