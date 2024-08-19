package valast

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type profiler struct {
	stack                 []reflect.Value
	invertedStackMessages []string
}

func (p *profiler) push(v reflect.Value) {
	if p == nil {
		return
	}
	p.stack = append(p.stack, v)
}

func (p *profiler) pop(startTime time.Time) {
	if p == nil {
		return
	}
	d := time.Since(startTime)
	v := p.stack[len(p.stack)-1].Interface()
	p.stack = p.stack[:len(p.stack)-1]
	stackSize := len(p.stack)
	msg := fmt.Sprintf("%s%vns: %T\n", strings.Repeat("  ", stackSize), d.Nanoseconds(), v)
	p.invertedStackMessages = append(p.invertedStackMessages, msg)
}

func (p *profiler) dump() {
	if p == nil {
		return
	}
	fmt.Println("valast: profile")
	for i := len(p.invertedStackMessages) - 1; i > 0; i-- {
		fmt.Print(p.invertedStackMessages[i])
	}
}
