package scout

// contains checks if a string slice contains a given value.
func Contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// getPercentage calculates the percentage of x in relation to y.
func GetPercentage(x, y float64) float64 {
	if x == 0 {
		return 0
	}

	if y == 0 {
		return 0
	}

	return x * 100 / y
}
