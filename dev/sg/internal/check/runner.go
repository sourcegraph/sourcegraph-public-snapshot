package check

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Runner[Args any] struct {
	in         io.Reader
	out        *std.Output
	categories []Category[Args]

	renderDescription func(*std.Output)
}

func NewRunner[Args any](in io.Reader, out *std.Output, categories []Category[Args]) *Runner[Args] {
	return &Runner[Args]{
		in:         in,
		out:        out,
		categories: categories,
	}
}

func (r *Runner[Args]) SetDescription(render func(out *std.Output)) {
	r.renderDescription = render
}

// Check executes all checks exactly once and exits.
func (r *Runner[Args]) Check(
	ctx context.Context,
	args Args,
) error {
	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	results := r.runAllCategoryChecks(ctx, args)
	if len(results.failed) > 0 {
		return errors.Newf("%d checks failed (%d skipped)", len(results.failed), len(results.skipped))
	}

	return nil
}

// Fix attempts to execute available fixes on checks that are not fulfilled.
func (r *Runner[Args]) Fix(
	ctx context.Context,
	args Args,
) error {
	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	// Get state
	results := r.runAllCategoryChecks(ctx, args)

	if len(results.failed) == 0 {
		return nil
	}

	r.out.WriteNoticef("Attempting to fix %d checks", len(results.failed))

	for _, i := range results.failed {
		category := r.categories[i]

		ok := r.fixCategoryAutomatically(ctx, &category, args, results)
		results.categories[category.Name] = !ok
	}

	// Report what is still bust
	failedCategories := []string{}
	for _, category := range r.categories {
		for _, c := range category.Checks {
			if !c.IsMet() {
				failedCategories = append(failedCategories, category.Name)
				break
			}
		}
	}
	if len(failedCategories) > 0 {
		return errors.Newf("Some categories are still unsatisfied: %s", strings.Join(failedCategories, ", "))
	}

	return nil
}

// Interactive runs both checks and fixes in an interactive manner, prompting the user for
// input.
func (r *Runner[Args]) Interactive(
	ctx context.Context,
	args Args,
) error {
	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	// Keep interactive runner up until all issues are fixed.
	results := &runAllCategoryChecksResult{
		failed: []int{1}, // initialize, this gets reset immediately
	}
	for len(results.failed) != 0 {
		r.out.Output.ClearScreen()

		// Get args when we run
		results = r.runAllCategoryChecks(ctx, args)
		if len(results.failed) == 0 {
			break
		}

		r.out.WriteWarningf("Some checks failed. Which one do you want to fix?")

		idx, err := r.getNumberOutOf(results.all)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		selectedCategory := r.categories[idx]

		r.out.ClearScreen()

		err = r.presentFailedCategoryWithOptions(ctx, idx, &selectedCategory, args, results)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}

	return nil
}

// runAllCategoryChecksResult provides a summary of categories checks results.
type runAllCategoryChecksResult struct {
	all     []int
	failed  []int
	skipped []int

	// Indicates whether each category succeeded or not
	categories map[string]bool
}

// runAllCategoryChecks is the main entrypoint for running the checks in this runner.
func (r *Runner[Args]) runAllCategoryChecks(ctx context.Context, args Args) *runAllCategoryChecksResult {
	if r.renderDescription != nil {
		r.renderDescription(r.out)
	}

	var results runAllCategoryChecksResult
	for i := range r.categories {
		results.failed = append(results.failed, i)
		results.all = append(results.all, i)
	}

	results.categories = make(map[string]bool)
	for _, category := range r.categories {
		results.categories[category.Name] = false
	}

	for i, category := range r.categories {
		idx := i + 1

		if err := category.CheckEnabled(ctx, args); err != nil {
			r.out.WriteSkippedf("%d. %s %s[SKIPPED: %s]%s",
				idx, category.Name, output.StyleBold, err.Error(), output.StyleReset)
			results.skipped = append(results.skipped, idx)
			results.failed = removeEntry(results.failed, i)
			continue
		}

		pending := r.out.Pending(output.Styledf(output.StylePending, "%d. %s - Determining status...", idx, category.Name))
		cio := IO{
			Input:  r.in,
			Writer: pending,
		}

		// Validate checks
		var failures []error
		for _, c := range category.Checks {
			if !c.IsEnabled(ctx, cio, args) {
				continue
			}

			if err := c.RunCheck(ctx, cio, args); err != nil {
				failures = append(failures, err)
			}
		}

		// Report results
		succeeded := len(failures) == 0
		results.categories[category.Name] = succeeded
		if succeeded {
			// If success, progress messages are not important
			pending.Destroy()
			r.out.WriteSuccessf("%d. %s", idx, category.Name)
			results.failed = removeEntry(results.failed, i)
		} else {
			// If failures, progress messages are important, so don't destroy
			for _, failure := range failures {
				pending.WriteLine(output.Styled(output.StyleWarning, failure.Error()))
			}
			pending.Complete(output.Linef(output.EmojiWarning, output.StyleFailure, "%d. %s", idx, category.Name))
		}
	}

	if len(results.failed) == 0 {
		if len(results.skipped) == 0 {
			r.out.Write("")
			r.out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
		} else {
			r.out.Write("")
			r.out.WriteWarningf("Some checks were skipped.")
		}
	}

	return &results
}

func removeEntry(s []int, val int) (result []int) {
	for _, e := range s {
		if e != val {
			result = append(result, e)
		}
	}
	return result
}

func (r *Runner[Args]) getNumberOutOf(numbers []int) (int, error) {
	var strs []string
	var idx = make(map[int]struct{})
	for _, num := range numbers {
		strs = append(strs, fmt.Sprintf("%d", num+1))
		idx[num+1] = struct{}{}
	}

	for {
		r.out.Writef("[%s]: ", strings.Join(strs, ","))
		var num int
		_, err := fmt.Fscan(r.in, &num)
		if err != nil {
			return 0, err
		}

		if _, ok := idx[num]; ok {
			return num - 1, nil
		}
		r.out.Writef("%d is an invalid choice :( Let's try again?\n", num)
	}
}

func (r *Runner[Args]) presentFailedCategoryWithOptions(ctx context.Context, categoryIdx int, category *Category[Args], args Args, results *runAllCategoryChecksResult) error {
	r.printCategoryHeaderAndDependencies(categoryIdx+1, category)
	fixableCategory := category.HasFixable()

	choices := map[int]string{1: "I want to fix these manually"}
	if fixableCategory {
		choices[2] = "I'm feeling lucky. You try fixing all of it for me."
		choices[3] = "Go back"
	} else {
		choices[2] = "Go back"
	}

	choice, err := r.getChoice(choices)
	if err != nil {
		return err
	}

	switch choice {
	case 1:
		err = r.fixCategoryManually(ctx, categoryIdx, category, args)
	case 2:
		if fixableCategory {
			r.out.ClearScreen()
			if !r.fixCategoryAutomatically(ctx, category, args, results) {
				err = errors.Newf("%s: failed to fix category automatically", category.Name)
			}
		}
	case 3:
		return nil
	}
	return err
}

func (r *Runner[Args]) printCategoryHeaderAndDependencies(categoryIdx int, category *Category[Args]) {
	r.out.WriteLine(output.Linef(output.EmojiLightbulb, output.CombineStyles(output.StyleSearchQuery, output.StyleBold), "%d. %s", categoryIdx, category.Name))
	r.out.Write("")
	r.out.Write("Checks:")

	for i, dep := range category.Checks {
		idx := i + 1
		if dep.IsMet() {
			r.out.WriteSuccessf("%d. %s", idx, dep.Name)
		} else {
			if dep.checkErr != nil {
				r.out.WriteFailuref("%d. %s: %s", idx, dep.Name, dep.checkErr)
			} else {
				r.out.WriteFailuref("%d. %s: %s", idx, dep.Name, "check failed")
			}
		}
	}
}

func (r *Runner[Args]) getChoice(choices map[int]string) (int, error) {
	for {
		r.out.Write("")
		r.out.WriteNoticef("What do you want to do?")

		for i := 0; i < len(choices); i++ {
			num := i + 1
			desc, ok := choices[num]
			if !ok {
				return 0, errors.Newf("internal error: %d not found in provided choices", i)
			}
			r.out.Writef("%s[%d]%s: %s", output.StyleBold, num, output.StyleReset, desc)
		}

		fmt.Printf("Enter choice: ")

		var s int
		_, err := fmt.Scan(&s)
		if err != nil {
			return 0, err
		}

		if _, ok := choices[s]; ok {
			return s, nil
		}
		r.out.WriteFailuref("Invalid choice")
	}
}

func (r *Runner[Args]) fixCategoryManually(ctx context.Context, categoryIdx int, category *Category[Args], args Args) error {
	for {
		toFix := []int{}

		for i, dep := range category.Checks {
			if dep.IsMet() {
				continue
			}

			toFix = append(toFix, i)
		}

		if len(toFix) == 0 {
			break
		}

		var idx int

		if len(toFix) == 1 {
			idx = toFix[0]
		} else {
			r.out.WriteNoticef("Which one do you want to fix?")
			var err error
			idx, err = r.getNumberOutOf(toFix)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}

		dep := category.Checks[idx]

		r.out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "%s", dep.Name))
		r.out.Write("")

		if dep.checkErr != nil {
			r.out.WriteLine(output.Styledf(output.StyleBold, "Encountered the following error:\n\n%s%s\n", output.StyleReset, dep.checkErr))
		}

		r.out.WriteLine(output.Styled(output.StyleBold, "How to fix:"))

		r.out.WriteMarkdown(dep.Description)

		pending := r.out.Pending(output.Styled(output.StylePending, "Determining status..."))
		for _, dep := range category.Checks {
			// update check state
			_ = dep.RunCheck(ctx, IO{
				Input:  r.in,
				Writer: pending,
			}, args)
		}
		pending.Destroy()

		r.printCategoryHeaderAndDependencies(categoryIdx, category)
	}

	return nil
}

func (r *Runner[Args]) fixCategoryAutomatically(ctx context.Context, category *Category[Args], args Args, results *runAllCategoryChecksResult) (ok bool) {
	pending := r.out.Pending(output.Styledf(output.StylePending, "Trying my hardest to fix %q automatically...", category.Name))
	cio := IO{
		Input:  r.in,
		Writer: pending,
	}

	if err := category.CheckEnabled(ctx, args); err != nil {
		pending.Complete(output.Styledf(output.StyleFailure, "%s - Skipped: %s", category.Name, err.Error()))
		return true
	}

	// If nothing in this category is fixable, we are done
	if !category.HasFixable() {
		pending.Complete(output.Styledf(output.StyleFailure, "%s - Cannot be fixed automatically.", category.Name))
		return false
	}

	// Only run if dependents are fixed
	var unmetDependencies []string
	for _, d := range category.DependsOn {
		if met, exists := results.categories[d]; !exists {
			pending.Complete(output.Styledf(output.StyleFailure, "%s - Required check category %q not found", category.Name, d))
			return false
		} else if !met {
			unmetDependencies = append(unmetDependencies, d)
		}
	}
	if len(unmetDependencies) > 0 {
		pending.Complete(output.Styledf(output.StyleFailure, "%s - Required dependencies %s not met.",
			category.Name, strings.Join(unmetDependencies, ", ")))
		return false
	}

	// now go through the real dependencies
	var fixFailed bool
	for _, c := range category.Checks {
		// If category is fixed, we are good to go
		if c.IsMet() {
			continue
		}

		if !c.IsEnabled(ctx, cio, args) {
			continue
		}

		// Otherwise, check if this is fixable at all
		if c.Fix == nil {
			pending.WriteLine(output.Styledf(output.StyleGrey, "%s cannot be fixed automatically.", c.Name))
			continue
		}

		// Attempt fix. Get new args because things might have changed due to another
		// fix being run.
		if err := c.Fix(ctx, cio, args); err != nil {
			pending.WriteLine(output.Styledf(output.StyleWarning, "Failed to fix %s: %s", c.Name, err.Error()))
			fixFailed = true
			continue
		}

		pending.WriteLine(output.Styledf(output.StylePending, "Ran fix for %s!", c.Name))

		// Check is the fix worked
		if err := c.RunCheck(ctx, cio, args); err != nil {
			pending.WriteLine(output.Styledf(output.StyleWarning, "Check %s still failing: %s",
				category.Name, c.Name, err.Error()))
			fixFailed = true
		}
	}

	if fixFailed {
		pending.Complete(output.Styledf(output.StyleFailure, "%s - Fixes failed", category.Name))
	} else {
		pending.Complete(output.Styledf(output.StyleSuccess, "%s - Done", category.Name))
	}

	return fixFailed
}

func waitForReturn() { fmt.Scanln() }
