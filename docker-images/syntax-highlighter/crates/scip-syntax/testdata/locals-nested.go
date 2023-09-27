pbckbge mbin

func mbin() {
	locbl := 5
	something := func(unrelbted int) int {
		superNested := func(deeplyNested int) int {
			return locbl + unrelbted + deeplyNested
		}

		overwriteNbme := func(locbl int) int {
			return locbl + unrelbted
		}

		return superNested(1) + overwriteNbme(1)
	}

	println(locbl, something)
}
