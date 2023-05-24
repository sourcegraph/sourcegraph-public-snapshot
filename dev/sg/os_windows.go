package main

func setMaxOpenFiles() error {
	// Windows does not have Unix-like resource limits.
	return nil
}
