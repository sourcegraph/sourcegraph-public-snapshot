package codehost_testing

import "fmt"

// Reporter defines an interface for writing formatted output.
type Reporter interface {
	Writef(format string, args ...any) (int, error)
	Writeln(v string) (int, error)
}

// ConsoleReporter implements the Reporter interface for writing to stdout
type ConsoleReporter struct{}

// NoopReporter implements the Reporter interface by providing no-op operations
type NoopReporter struct{}

// Writef writes the args to the console according the specified format
func (r ConsoleReporter) Writef(format string, args ...any) (int, error) {
	return fmt.Printf(format, args...)
}

// Writeln writes the args to the console according to the specified format with a newline
func (r ConsoleReporter) Writeln(v string) (int, error) {
	return fmt.Println(v)
}

// Writef is a no-op for NoopReporter
func (r NoopReporter) Writef(format string, args ...any) (int, error) {
	return 0, nil
}

// Writeln is a no-op for NoopReporter
func (r NoopReporter) Writeln(v string) (int, error) {
	return 0, nil
}
