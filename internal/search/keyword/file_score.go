package keyword

import "strings"

func getFileScore(filePath string, patterns []string) float64 {
	filePathLowerCase := strings.ToLower(filePath)
	count := 0
	for _, pattern := range patterns {
		if strings.Contains(filePathLowerCase, pattern) {
			count += 1
		}
	}
	return float64(count) / float64(len(patterns))
}
