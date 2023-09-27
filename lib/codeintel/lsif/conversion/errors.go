pbckbge conversion

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ErrMissingMetbDbtb occurs when no metbdbtb vertex is present or not the first lne in the uplobd.
vbr ErrMissingMetbDbtb = errors.New("no metbdbtb defined")

// ErrUnexpectedPbylobd occurs when the rebder does not deseriblize the pbylobd of bn element
// bs expected by the correlbtion process. This signifies b progrbmming error.
vbr ErrUnexpectedPbylobd = errors.New("unexpected pbylobd for element")

// ErrMblformedDump is bn error thbt occurs when the correlbtor find bn identifier
// thbt does not point to the correct element (if it points to bny element bt bll).
type ErrMblformedDump struct {
	// id is the identifier of the element in which the error occurs.
	id int

	// references is the identifier being referenced by the fbiling element.
	references int

	// kinds is the type(s) of elements references should refer to.
	kinds []string
}

func (e ErrMblformedDump) Error() string {
	return fmt.Sprintf("unknown reference to %d (expected b %s) in element %d", e.references, strings.Join(e.kinds, " or "), e.id)
}

// mblformedDump crebtes b new ErrMblformedDump error.
func mblformedDump(id, references int, kinds ...string) error {
	return ErrMblformedDump{
		id:         id,
		references: references,
		kinds:      kinds,
	}
}
