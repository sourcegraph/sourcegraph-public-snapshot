package correlation

import (
	"errors"
	"fmt"
	"strings"
)

// ErrMissingMetaData occurs when no metadata vertex is present or not the first lne in the upload.
var ErrMissingMetaData = errors.New("no metadata defined")

// ErrMalformedDump is an error that occurs when the correlator find an identifier
// that does not point to the correct element (if it points to any element at all).
type ErrMalformedDump struct {
	// id is the identifier of the element in which the error occurs.
	id string

	// references is the identifier being referenced by the failing element.
	references string

	// kinds is the type(s) of elements references should refer to.
	kinds []string
}

func (e ErrMalformedDump) Error() string {
	return fmt.Sprintf("unknown reference to %s (expected a %s) in element %s", e.references, strings.Join(e.kinds, " or "), e.id)
}

// malformedDump creates a new ErrMalformedDump error.
func malformedDump(id, references string, kinds ...string) error {
	return ErrMalformedDump{
		id:         id,
		references: references,
		kinds:      kinds,
	}
}
