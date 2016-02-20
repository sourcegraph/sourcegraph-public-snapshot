// Provides convience methods to print information to color-aware terminal
package colorable

import (
	"fmt"
)

// Mimics fmt.Print
func Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(Stdout, a...)
}

// Mimics fmt.Printf
func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(Stdout, format, a...)
}

// Mimics fmt.Println
func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(Stdout, a...)
}
