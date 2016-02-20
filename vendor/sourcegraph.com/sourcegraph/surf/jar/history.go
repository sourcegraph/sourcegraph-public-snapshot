package jar

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

// State represents a point in time.
type State struct {
	Request  *http.Request
	Response *http.Response
	Dom      *goquery.Document
}

// NewHistoryState creates and returns a new *State type.
func NewHistoryState(req *http.Request, resp *http.Response, dom *goquery.Document) *State {
	return &State{
		Request:  req,
		Response: resp,
		Dom:      dom,
	}
}

// History is a type that records browser state.
type History interface {
	Len() int
	Push(p *State) int
	Pop() *State
	Top() *State
}

// Node holds stack values and points to the next element.
type Node struct {
	Value *State
	Next  *Node
}

// MemoryHistory is an in-memory implementation of the History interface.
type MemoryHistory struct {
	top  *Node
	size int
}

// NewMemoryHistory creates and returns a new *StateHistory type.
func NewMemoryHistory() *MemoryHistory {
	return &MemoryHistory{}
}

// Len returns the number of states in the history.
func (his *MemoryHistory) Len() int {
	return his.size
}

// Push adds a new State at the front of the history.
func (his *MemoryHistory) Push(p *State) int {
	his.top = &Node{p, his.top}
	his.size++
	return his.size
}

// Pop removes and returns the State at the front of the history.
func (his *MemoryHistory) Pop() *State {
	if his.size > 0 {
		value := his.top.Value
		his.top = his.top.Next
		his.size--
		return value
	}

	return nil
}

// Top returns the State at the front of the history without removing it.
func (his *MemoryHistory) Top() *State {
	if his.size == 0 {
		return nil
	}
	return his.top.Value
}
