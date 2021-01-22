package inference

func containsIndex(haystack []string, needle string) (bool, int) {
	for i, value := range haystack {
		if value == needle {
			return true, i
		}
	}

	return false, -1
}

func contains(haystack []string, needle string) bool {
	ok, _ := containsIndex(haystack, needle)
	return ok
}
