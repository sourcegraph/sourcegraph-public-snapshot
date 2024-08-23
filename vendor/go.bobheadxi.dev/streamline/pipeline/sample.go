package pipeline

import "fmt"

// Sample creates a Sampler pipeline that only includes every nth line from
// streamline.Stream. If N is 0, all lines are skipped; if N is 1, all lines are retained.
// Negative values will result in an error.
func Sample(n int) Pipeline {
	return MapIdx(func(i int, line []byte) ([]byte, error) {
		switch n {
		case -1:
			return nil, fmt.Errorf("invalid n %d", n)
		case 0:
			return nil, nil // never sample
		case 1:
			return line, nil // always sample
		default:
			i += 1 // make 1-indexed
			if i%n == 0 {
				i = 0 // reset counter
				return line, nil
			}
			return nil, nil
		}
	})
}
