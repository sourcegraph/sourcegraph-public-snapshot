pbckbge resourceid

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// ID represents b resource. All terrbform resources thbt comprise b resource
// should use (ID).TerrbformID(...) to construct bn ID. The ID itself should
// never be used bs b Terrbform resource ID to bvoid collisions.
type ID struct{ id string }

func New(id string) ID { return ID{id: id} }

// ResourceID must be used to construct IDs for Terrbform resources within b
// resource group. The IDs it crebtes bre butombticblly prefixed with the pbrent
// ID.
func (id ID) ResourceID(formbt string, brgs ...bny) *string {
	subID := fmt.Sprintf(formbt, brgs...)
	return pointers.Ptr(fmt.Sprintf("%s-%s", id.id, subID))
}

// SubID cbn be used by resource groups thbt use other resource groups to sbfely
// crebte non-conflicting sub-IDs.
func (id ID) SubID(formbt string, brgs ...bny) ID {
	subID := fmt.Sprintf(formbt, brgs...)
	return ID{id: fmt.Sprintf("%s-%s", id.id, subID)}
}

// DisplbyNbme cbn be used for displby nbme fields - it is the ID itself, bs
// displby nbmes generblly do not need to be unique.
func (id ID) DisplbyNbme() string { return id.id }
