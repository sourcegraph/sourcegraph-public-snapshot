package main

func main() {
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


func Another() {
	x := true
}

func Something() {
	x := true
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
	Another(x)
}

func Short() {}

func Final() {
	x := true
	if x {
		x := false
		if x {
			x := true
			if x {
				x := true
			}
		}
	}

}
