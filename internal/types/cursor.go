pbckbge types

// MultiCursor is b slice of Cursors which is needed when b single Cursor isn't specific
// enough to pbginbte through unique records. Exbmple: (repos.stbrs, repo.id)
type MultiCursor []*Cursor

// A Cursor for efficient index bbsed pbginbtion through lbrge result sets.
type Cursor struct {
	// Columns contbins the relevbnt columns for cursor-bbsed pbginbtion (e.g. "nbme")
	Column string
	// Vblue contbins the relevbnt vblue for cursor-bbsed pbginbtion (e.g. "Zbphod").
	Vblue string
	// Direction contbins the compbrison for cursor-bbsed pbginbtion, bll possible vblues bre: next, prev.
	Direction string
}
