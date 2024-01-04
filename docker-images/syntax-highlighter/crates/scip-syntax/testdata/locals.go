package main

// Not actually local
var local = 10

func main() {
	local = 20
	local := 5
	something := func(unrelated int) int {
		superNested := func(deeplyNested int) int {
			return local + unrelated + deeplyNested
		}

		overwriteName := func(local int) int {
			return local + unrelated
		}

		return superNested(1) + overwriteName(1)
	}

	println(local, something)
}

func Another(local int) int {
	if local := 9; local < 0 {
		fmt.Println(local, "is negative")
	} else if local < 10 {
		fmt.Println(local, "has 1 digit")
	} else {
		fmt.Println(local, "has multiple digits")
	}
	return local
}
