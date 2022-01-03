package highlight

import tree_sitter "github.com/smacker/go-tree-sitter"

type AllOrderIterator struct {
	visitStack []VisitInfo
	visitor    TreeVisitor
}

type TreeVisitor interface {
	// startVisit is called when a node is first visited, before any of its children.
	startVisit(n *tree_sitter.Node) error
	// afterVisitingChild is called after the n-th child of parentNode has
	// been visited.
	//
	// This method is not called if parentNode doesn't have any children.
	afterVisitingChild(parentNode *tree_sitter.Node, visitedChildIndex int) error
	// endVisit is called after the last child of parentNode has been visited.
	//
	// Technically, this is redundant since one can check whether the child was the
	// last one or not in afterVisitingChild, but this is provided for ergonomics.
	endVisit(n *tree_sitter.Node) error
	// Special handling for anonymous nodes. Anonymous nodes' children, if any,
	// are not visited separately in AllOrderIterator's traversal, since these
	// usually correspond to literals in the grammar, such as keywords and
	// punctuation.
	//
	// NOTE: We use 'anonymous' instead of 'unnamed' since 'anonymous' is the
	// official tree-sitter terminology.
	// (https://tree-sitter.github.io/tree-sitter/using-parsers#named-vs-anonymous-nodes)
	visitAnonymous(n *tree_sitter.Node) error
}

// Invariant: -1 <= next <= node.ChildCount()
type VisitInfo struct {
	node *tree_sitter.Node
	next int
}

// NewIterator takes a node and mode (DFS/BFS) and returns iterator over children of the node
func NewAllOrderIterator(root *tree_sitter.Node, visitor TreeVisitor) AllOrderIterator {
	return AllOrderIterator{
		[]VisitInfo{{node: root, next: -1}},
		visitor,
	}
}

/*

           A
           |
           B
          / \
         C   D

                                       [{A, -1}]
                      [startVisit A]-> [{A, 0}, {B, -1}]
                      [startVisit B]-> [{A, 0}, {B, 0}, {C, -1}]
                      [startVisit C]-> [{A, 0}, {B, 0}, {C, 0}]
                        [endVisit C]-> [{A, 0}, {B, 1}]
            [afterVisitingChild B 0]-> [{A, 0}, {B, 1}, {D, -1}]
                      [startVisit D]-> [{A, 0}, {B, 1}, {D, 0}]
                        [endVisit D]-> [{A, 0}, {B, 2}
[afterVisitingChild B 1, endVisit B]-> [{A, 1}]
[afterVisitingChild A 0, endVisit A]-> []

*/

func (i *AllOrderIterator) popEnd() {
	i.visitStack = i.visitStack[:len(i.visitStack)-1]
	if len(i.visitStack) > 0 {
		i.visitStack[len(i.visitStack)-1].next++
	}
}

func (i *AllOrderIterator) VisitTree() (VisitInfo, error) {
	for {
		if len(i.visitStack) == 0 {
			return VisitInfo{}, nil
		}
		last := i.visitStack[len(i.visitStack)-1]
		if last.node == nil {
			i.popEnd()
			continue
		}
		if !last.node.IsNamed() {
			if err := i.visitor.visitAnonymous(last.node); err != nil {
				return last, err
			}
			i.popEnd()
			continue
		}
		if last.node.IsNull() || last.node.IsMissing() {
			panic("TODO: [Varun] Add handling for null and missing nodes.")
		}
		switch last.next {
		case -1:
			if err := i.visitor.startVisit(last.node); err != nil {
				return last, err
			}
			i.visitStack[len(i.visitStack)-1].next = 0
			if last.node.ChildCount() > 0 {
				i.visitStack = append(i.visitStack, VisitInfo{last.node.Child(0), -1})
			}
			continue
		case int(last.node.ChildCount()):
			if last.node.ChildCount() > 0 {
				childIdx := int(last.node.ChildCount()) - 1
				if err := i.visitor.afterVisitingChild(last.node, childIdx); err != nil {
					return VisitInfo{last.node, childIdx}, err
				}
			}
			if err := i.visitor.endVisit(last.node); err != nil {
				return last, err
			}
			i.popEnd()
			continue
		}
		if last.next < 0 {
			panic("Found negative next value other than -1")
		}
		if last.next > 0 {
			if err := i.visitor.afterVisitingChild(last.node, last.next-1); err != nil {
				return VisitInfo{last.node, last.next - 1}, err
			}
		}
		i.visitStack = append(i.visitStack, VisitInfo{last.node.Child(last.next), -1})
		continue
	}
}
