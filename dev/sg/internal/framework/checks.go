package framework

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Report struct {
	// TODO
	Err error
}

type Checker[Arg any] interface {
	Check(context.Context, Arg) *Report
}

type FixableChecker[Arg any] interface {
	Checker[Arg]
	Fix(context.Context, Arg) *Report
}

type CheckerEnabler[Arg any] interface {
	Checker[Arg]
	Enabler[Arg]
}

type Category[Arg any, C Checker[Arg]] interface {
	Name() string
	Checks() []C
}

type CategoryEnabler[Arg any, C Checker[Arg]] interface {
	Enabler[Arg]
	Category[Arg, C]
}

type Enabler[Arg any] interface {
	Enabled(Arg) error
}

type RunMode int

const (
	Check = 1 << iota
	Fix
	Interactive
)

type runner[Arg any, C Checker[Arg]] struct {
	checks []C
}

func (r *runner[Arg, C]) RunInteractive(
	ctx context.Context,
	out *std.Output,
	argBuilder func() Arg,
	mode RunMode,
	categories []Category[Arg, C],
) error {
	all := []int{}
	failed := []int{}
	disabled := []int{}
	for i := range categories {
		failed = append(failed, i)
		all = append(all, i)
	}

	for len(failed) != 0 {
		out.ClearScreen()

		// TODO parameterize help text

		arg := argBuilder()

		for i, category := range categories {
			idx := i + 1

			if enabler, ok := category.(Enabler[Arg]); ok {
				if err := enabler.Enabled(arg); err != nil {
					std.Out.WriteSkippedf("%d. %s %s[DISABLED: %s]%s",
						idx, category.Name(), output.StyleBold, err.Error(), output.StyleReset)
					disabled = append(disabled, idx)
					failed = removeEntry(failed, i)
					continue
				}
			}

			pending := std.Out.Pending(output.Styledf(output.StylePending, "%d. %s - Determining status...", idx, category.name))

			checks := category.Checks()
			reports := make([]*Report, len(checks))
			var hasFailedChecks bool
			for i, c := range checks {
				reports[i] = c.Check(ctx, arg)
				if reports[i].Err != nil {
					hasFailedChecks = true
				}
			}
			pending.Destroy()

			if !hasFailedChecks {
				out.WriteSuccessf("%d. %s", idx, category.Name())
				failed = removeEntry(failed, i)
			} else {
				// TODO write failure
				out.WriteFailuref("%d. %s", idx, category.Name())
			}
		}

		if len(failed) == 0 {
			if len(disabled) == 0 {
				out.Write("")
				out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
			} else {
				out.Write("")
				out.WriteWarningf("Some checks were disabled.")
				// std.Out.WriteSuggestionf("Restart 'sg setup' in the 'sourcegraph' repository to continue.")
			}

			return nil
		}

		std.Out.Write("")

		std.Out.WriteWarningf("Some checks failed. Which one do you want to fix?")

		idx, err := getNumberOutOf(all)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		selectedCategory := categories[idx]

		std.Out.ClearScreen()

		// TODO
		// err = presentFailedCategoryWithOptions(ctx, idx, &selectedCategory)
		// if err != nil {
		// 	if err == io.EOF {
		// 		return nil
		// 	}
		// 	return err
		// }
	}

	return nil
}

type sgSetupArgType struct {
	Teammate bool
	InRepo   bool
}

func removeEntry(s []int, val int) (result []int) {
	for _, e := range s {
		if e != val {
			result = append(result, e)
		}
	}
	return result
}

func getNumberOutOf(numbers []int) (int, error) {
	var strs []string
	var idx = make(map[int]struct{})
	for _, num := range numbers {
		strs = append(strs, fmt.Sprintf("%d", num+1))
		idx[num+1] = struct{}{}
	}

	for {
		fmt.Printf("[%s]: ", strings.Join(strs, ","))
		var num int
		_, err := fmt.Scan(&num)
		if err != nil {
			return 0, err
		}

		if _, ok := idx[num]; ok {
			return num - 1, nil
		}
		fmt.Printf("%d is an invalid choice :( Let's try again?\n", num)
	}
}
