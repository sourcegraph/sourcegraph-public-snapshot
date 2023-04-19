package util

import "fmt"

// FormatPreKey returns the key for the pre step with the given index.
func FormatPreKey(index int) string {
	return fmt.Sprintf("step.%d.pre", index)
}

// FormatRunKey returns the key for the run step with the given index.
func FormatRunKey(index int) string {
	return fmt.Sprintf("step.%d.run", index)
}

// FormatPostKey returns the key for the post step with the given index.
func FormatPostKey(index int) string {
	return fmt.Sprintf("step.%d.post", index)
}
