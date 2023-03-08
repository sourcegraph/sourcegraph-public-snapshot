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
