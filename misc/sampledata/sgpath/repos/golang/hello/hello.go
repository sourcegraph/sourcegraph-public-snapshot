// Package hello constructs friendly greetings.
package hello

import "fmt"

// World returns a "Hello, world!" greeting.
func World() string {
	return "Hello, world!"
}

// Name returns a greeting for the named person. If name is empty, it
// returns a generic greeting.
func Name(name string) string {
	if name == "" {
		return "Hello!"
	}
	return fmt.Sprintf("Hello, %s!", name)
}
