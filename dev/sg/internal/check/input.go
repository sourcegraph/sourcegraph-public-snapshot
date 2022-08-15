package check

import (
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func getChoice(in io.Reader, out *std.Output, choices map[int]string) (int, error) {
	for {
		out.Write("")
		out.WriteNoticef("What do you want to do?")

		for i := 0; i < len(choices); i++ {
			num := i + 1
			desc, ok := choices[num]
			if !ok {
				return 0, errors.Newf("internal error: %d not found in provided choices", i)
			}
			out.Writef("%s[%d]%s: %s", output.StyleBold, num, output.StyleReset, desc)
		}

		out.Promptf("Enter choice:")
		var s int
		if _, err := fmt.Fscan(in, &s); err != nil {
			return 0, err
		}

		if _, ok := choices[s]; ok {
			return s, nil
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

func waitForReturn(in io.Reader) {
	fmt.Fscanln(in)
}
