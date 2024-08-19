package ast

// ParseNode represents a function to parse ast nodes.
type ParseNode func(p *Parser) (*Node, error)

// Node is a simple node in a tree with double linked lists instead of slices to
// keep track of its siblings and children. A node is either a value or a
// parent node.
type Node struct {
	// Type of the node.
	Type int
	// TypeStrings contains all the string representations of the available types.
	TypeStrings []string
	// Value of the node. Only possible if it has no children.
	Value string

	// Parent is the parent node.
	Parent *Node
	// PreviousSibling is the previous sibling of the node.
	PreviousSibling *Node
	// NextSibling is the next sibling of the node.
	NextSibling *Node
	// FirstChild is the first child of the node.
	FirstChild *Node
	// LastChild is the last child of the node.
	LastChild *Node
}

// TypeString returns the strings representation of the type. Same as TypeStrings[Type]. Returns "UNKNOWN" if not
// string representation is found or len(TypeStrings) == 0.
func (n *Node) TypeString() string {
	if 0 <= n.Type && n.Type < len(n.TypeStrings) {
		return n.TypeStrings[n.Type]
	}
	return "UNKNOWN"
}

// IsParent returns whether the node has children and thus is not a value node.
func (n *Node) IsParent() bool {
	return n.FirstChild != nil
}

// Children returns all the children of the node.
func (n *Node) Children() []*Node {
	if n.FirstChild == nil {
		return nil //  Node has no children.
	}
	var cs []*Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		cs = append(cs, c)
	}
	return cs
}

// Remove removes itself from the tree.
func (n *Node) Remove() *Node {
	if n.Parent != nil {
		// Set the first child to the next.
		if n.Parent.FirstChild == n {
			n.Parent.FirstChild = n.NextSibling
		}
		// Set the last child to the previous.
		if n.Parent.LastChild == n {
			n.Parent.LastChild = n.PreviousSibling
		}
		n.Parent = nil
	}

	if n.PreviousSibling != nil {
		// Set the next sibling of the previous sibling to the next.
		n.PreviousSibling.NextSibling = n.NextSibling
	}
	if n.NextSibling != nil {
		// Set the previous sibling of the next sibling to the previous.
		n.NextSibling.PreviousSibling = n.PreviousSibling
	}
	n.NextSibling = nil
	n.PreviousSibling = nil
	return n
}

func (n *Node) Adopt(other *Node) {
	if other.FirstChild == nil {
		// Nothing to adapt.
		return
	}
	n.SetLast(other.FirstChild.Remove())
	n.Adopt(other)
}

// SetPrevious inserts the given node as the previous sibling.
func (n *Node) SetPrevious(sibling *Node) {
	sibling.Remove()
	sibling.Parent = n.Parent
	if n.PreviousSibling != nil {
		// Already has a sibling.
		// 1. Update next of previous node.
		// 2. Assign sibling as previous.
		// 3. Copy over previous of node.
		// 4. Add node as next.
		n.PreviousSibling.NextSibling = sibling     // (1)
		n.PreviousSibling = sibling                 // (2)
		sibling.PreviousSibling = n.PreviousSibling // (3)
		sibling.NextSibling = n                     // (4)
	}
	// Does not have a previous sibling yet.
	// 1. Reference each other.
	// 2. Update references of parent.
	n.PreviousSibling = sibling // (1)
	sibling.NextSibling = n
	if n.Parent != nil { // (2)
		n.Parent.FirstChild = sibling
	}
}

// SetNext inserts the given node as the next sibling.
func (n *Node) SetNext(sibling *Node) {
	sibling.Remove()
	sibling.Parent = n.Parent
	// (a) <-> (b) | a.AddSibling(b)
	// (a) <-> (c) <-> (b)
	if n.NextSibling != nil {
		// Already has a sibling.
		// 1. Update previous of next node.
		// 2. Assign sibling as next.
		// 3. Copy over next of node.
		// 4. Add node as previous.
		n.NextSibling.PreviousSibling = sibling // (1)
		n.NextSibling = sibling                 // (2)
		sibling.NextSibling = n.NextSibling     // (3)
		sibling.PreviousSibling = n             // (4)
		return
	}
	// Does not have a next sibling yet.
	// 1. Reference each other.
	// 2. Update references of parent.
	sibling.PreviousSibling = n // (1)
	n.NextSibling = sibling
	if n.Parent != nil { // (2)
		n.Parent.LastChild = sibling
	}
}

// SetFirst inserts the given node as the first child of the node.
func (n *Node) SetFirst(child *Node) {
	child.Remove()
	// Set the first child.
	if n.FirstChild != nil {
		n.FirstChild.SetPrevious(child)
		return
	}
	// No children present.
	child.Parent = n
	n.FirstChild = child
	n.LastChild = child
}

// SetLast inserts the given node as the last child of the node.
func (n *Node) SetLast(child *Node) {
	child.Remove()
	// Set the last child.
	if n.FirstChild != nil {
		n.LastChild.SetNext(child)
		return
	}
	// No children present.
	child.Parent = n
	n.FirstChild = child
	n.LastChild = child
}
