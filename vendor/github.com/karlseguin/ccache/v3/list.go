package ccache

type List[T any] struct {
	Head *Node[T]
	Tail *Node[T]
}

func NewList[T any]() *List[T] {
	return &List[T]{}
}

func (l *List[T]) Remove(node *Node[T]) {
	next := node.Next
	prev := node.Prev

	if next == nil {
		l.Tail = node.Prev
	} else {
		next.Prev = prev
	}

	if prev == nil {
		l.Head = node.Next
	} else {
		prev.Next = next
	}
	node.Next = nil
	node.Prev = nil
}

func (l *List[T]) MoveToFront(node *Node[T]) {
	l.Remove(node)
	l.nodeToFront(node)
}

func (l *List[T]) Insert(value T) *Node[T] {
	node := &Node[T]{Value: value}
	l.nodeToFront(node)
	return node
}

func (l *List[T]) nodeToFront(node *Node[T]) {
	head := l.Head
	l.Head = node
	if head == nil {
		l.Tail = node
		return
	}
	node.Next = head
	head.Prev = node
}

type Node[T any] struct {
	Next  *Node[T]
	Prev  *Node[T]
	Value T
}
