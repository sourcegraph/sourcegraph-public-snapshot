package std

import (
	"bufio"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

// PromptAndScan is a helper that renders the prompt into out and scans for
// subsequent input up to a newline. The return value indicates if a value was
// provided at all.
//
//	ok, err := std.PromptAndScan(std.Out, "Prompt:", &value)
//	if err != nil {
//		return err
//	} else if !ok {
//		return errors.New("response is required")
//	}
func PromptAndScan(out *Output, prompt string, result *string) (bool, error) {
	out.Promptf(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	*result = strings.TrimSpace(input)
	return *result != "", nil
}

// FancyPromptAndScan is a helper that renders the given fancy prompt into out and scans for the
// subsequent input up to a newline. The return value indicates if a value was provided at all
func FancyPromptAndScan(out *Output, prompt output.FancyLine, result *string) (bool, error) {
	out.FancyPrompt(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	*result = strings.TrimSpace(input)
	return *result != "", nil
}
