package example

import (
	f "fmt"
)

func Something() {
	x := true
	f.Println(x)
}

func Another() float64 { return 5 / 3 }

type MyThing struct{}

func (m *MyThing) DoSomething()    {}
func (m MyThing) DoSomethingElse() {}
