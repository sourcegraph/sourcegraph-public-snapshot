package main

func strp(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}

func intp(v *int) int {
	if v == nil {
		return 0
	}

	return *v
}
