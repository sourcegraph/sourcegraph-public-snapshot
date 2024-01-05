package check

import (
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func getChoice(in io.Reader, out *std.Output, choices []string) (string, error) {
	for {
		out.Write("")
		out.WriteNoticef("What do you want to do?")

		for i, choice := range choices {
			out.Writef("%s[%d]%s: %s", output.StyleBold, i+1, output.StyleReset, choice)
		}

		out.Promptf("Enter choice:")
		var s int
		if _, err := fmt.Fscan(in, &s); err != nil {
			return "", err
		}

		if s > 0 && s <= len(choices) {
			// decrement to zero indexing
			return choices[s-1], nil
		}
		out.WriteFailuref("Invalid choice")
	}
}

func getNumberOutOf(in io.Reader, out *std.Output, numbers []int) (int, error) {
	var strs []string
	var idx = make(map[int]struct{})
	for _, num := range numbers {
		strs = append(strs, fmt.Sprintf("%d", num+1))
		idx[num+1] = struct{}{}
	}

	for {
		out.Promptf("[%s]:", strings.Join(strs, ","))
		var num int
		_, err := fmt.Fscan(in, &num)
		if err != nil {
			return 0, err
		}

		if _, ok := idx[num]; ok {
			return num - 1, nil
		}
		out.Writef("%d is an invalid choice :( Let's try again?\n", num)
	}
}

func getOptionOutOf(in io.Reader, out *std.Output, options []string) (string, error) {
	var optMap = make(map[string]struct{}, len(options))

	for _, opt := range options {
		optMap[opt] = struct{}{}
	}

	for {
		out.Promptf("[%s]:", strings.Join(options, ","))
		var choice string
		_, err := fmt.Fscan(in, &choice)
		if err != nil {
			return "", err
		}

		if _, ok := optMap[choice]; ok {
			return choice, nil
		}
		out.Writef("%s is an invalid choice :( Let's try again?\n", choice)
	}
}

func waitForReturn(in io.Reader) {
	fmt.Fscanln(in)
}
