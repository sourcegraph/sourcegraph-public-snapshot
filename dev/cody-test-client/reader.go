package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type readerFunc func(prompt string) string

func newStdinReader() readerFunc {
	reader := bufio.NewReader(os.Stdin)

	return func(prompt string) string {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		return strings.TrimSpace(input)
	}
}
