package main

import (
	"fmt"
)

func main() {
	// Variables
	var x int = 5
	y := 10
	// Constants
	const z = 15
	// Arrays
	array := [5]int{1, 2, 3, 4, 5}
	// Slices
	slice := array[1:3]
	slice = append(slice, 6)
	// Maps
	m := map[string]int{"foo": 42}
	// Structs
	type person struct {
		name string
		age  int
	}
	p := person{"Bob", 50}
	// Interfaces
	var i interface{} = p
	fmt.Println(i.(person).name)
	// Error handling
	if err := foo(); err != nil {
		fmt.Println(err)
	}
	// Functions
	defered()
	go concurrent()
	pointers()
	// Looping and branching
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue
		}
		fmt.Println(i)
		if i > 5 {
			break
		}
	}
	// Type conversions
	j := int8(x)
	// Packages
	math.MaxInt32
	// And more...
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

}
func foo() error {
	return fmt.Errorf("foo error")
}
func defered() {
	defer fmt.Println("deferred")
}
func concurrent() {
	go func() {
		fmt.Println("concurrent")
	}()
}
func pointers() {
	x := 5
	fmt.Println(&x) // print memory address
}

type Person struct {
	Name string
	Age  int
}
type Vehicle struct {
	Wheels int
	Owner  *Person
}
type Drivable interface {
	Wheels() int
}

func structExample() {
	p := Person{"Bob", 50}
	v := Vehicle{Wheels: 4, Owner: &p}
	var d Drivable = v
	fmt.Println(d.Wheels()) // 4
	v.Owner.Age = 51
	fmt.Println(p.Age) // 51
}

func Min[T Comparable](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func generics() {
	fmt.Println(Min[int](5, 10))       // 5
	fmt.Println(Min[string]("a", "b")) // "a"
}
