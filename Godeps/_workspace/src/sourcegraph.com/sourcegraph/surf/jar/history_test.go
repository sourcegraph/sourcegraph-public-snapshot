package jar

import (
	"github.com/headzoo/ut"
	"testing"
)

func TestMemoryHistory(t *testing.T) {
	ut.Run(t)
	stack := NewMemoryHistory()

	page1 := &State{}
	stack.Push(page1)
	ut.AssertEquals(1, stack.Len())
	ut.AssertEquals(page1, stack.Top())

	page2 := &State{}
	stack.Push(page2)
	ut.AssertEquals(2, stack.Len())
	ut.AssertEquals(page2, stack.Top())

	page := stack.Pop()
	ut.AssertEquals(page, page2)
	ut.AssertEquals(1, stack.Len())
	ut.AssertEquals(page1, stack.Top())

	page = stack.Pop()
	ut.AssertEquals(page, page1)
	ut.AssertEquals(0, stack.Len())
}
