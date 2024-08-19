package constructs


// In what order to return constructs.
type ConstructOrder string

const (
	// Depth-first, pre-order.
	ConstructOrder_PREORDER ConstructOrder = "PREORDER"
	// Depth-first, post-order (leaf nodes first).
	ConstructOrder_POSTORDER ConstructOrder = "POSTORDER"
)

