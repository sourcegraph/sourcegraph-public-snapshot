package main

import "fmt"

func main() {
	read := newStdinReader()
	capabilities := newCapabilities(read)

	for {
		name := read("Select a pipeline: ")
		if name == "" || name == "exit" || name == "quit" {
			break
		}

		if err := runPipeline(name, capabilities); err != nil {
			fmt.Printf("An error occurred: %s\n", err)
		}

		fmt.Printf("\n")
	}
}
