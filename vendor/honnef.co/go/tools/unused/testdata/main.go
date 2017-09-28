package main

func Fn1() {}
func Fn2() {} // MATCH /Fn2 is unused/

const X = 1 // MATCH /X is unused/

var Y = 2 // MATCH /Y is unused/

type Z struct{} // MATCH /Z is unused/

func main() {
	Fn1()
}
