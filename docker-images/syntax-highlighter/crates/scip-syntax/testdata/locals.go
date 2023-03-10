package main

func main() {
	local := true
	something := func(local int) int {
		return local
	}

	println(local, something)
}

func Another(local int) int {
	return local
}
