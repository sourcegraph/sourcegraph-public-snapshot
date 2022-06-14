package check

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"go.uber.org/atomic"

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

// NewRunner creates a Runner for executing checks and applying fixes in a variety of ways.
func NewRunner[Args any](in io.Reader, out *std.Output, categories []Category[Args]) *Runner[Args] {
	return &Runner[Args]{
		in:         in,
		out:        out,
		categories: categories,
	}
}

// SetDescription sets a description to render before core check loops, such as a massive
// ASCII art thing.
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

// Fix attempts to applies available fixes on checks that are not satisfied.
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
		// Nothing failed, we're good to go!
		return nil
	}

	r.out.WriteNoticef("Attempting to fix %d failed categories", len(results.failed))
	for _, i := range results.failed {
		category := r.categories[i]

		ok := r.fixCategoryAutomatically(ctx, i+1, &category, args, results)
		results.categories[category.Name] = ok
	}

	// Report what is still bust
	failedCategories := []string{}
	for c, ok := range results.categories {
		if ok {
			continue
		}
		failedCategories = append(failedCategories, fmt.Sprintf("%q", c))
	}
	if len(failedCategories) > 0 {
		return errors.Newf("Some categories are still unsatisfied: %s", strings.Join(failedCategories, ", "))
	}

	return nil
}

// Interactive runs both checks and fixes in an interactive manner, prompting the user for
// decisions about which fixes to apply.
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

		results = r.runAllCategoryChecks(ctx, args)
		if len(results.failed) == 0 {
			break
		}

		r.out.WriteWarningf("Some checks failed. Which one do you want to fix?")

		idx, err := getNumberOutOf(r.in, r.out, results.failed)
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

	statuses := []*output.StatusBar{}
	var checks float64
	for i, category := range r.categories {
		statuses = append(statuses, output.NewStatusBarWithLabel(fmt.Sprintf("%d. %s", i+1, category.Name)))
		checks += float64(len(category.Checks))
	}
	progress := r.out.ProgressWithStatusBars([]output.ProgressBar{{
		Label: "Running checks",
		Max:   checks,
	}}, statuses, nil)

	var (
		categoriesWg sync.WaitGroup
		skipped      = map[int]error{}

		// used for progress bar
		checksDone           atomic.Float64
		updateChecksProgress = func() {
			progress.SetValue(0, checksDone.Load()+1)
			checksDone.Add(1)
		}
	)
	for i, category := range r.categories {
		progress.StatusBarUpdatef(i, "Determining status...")

		if err := category.CheckEnabled(ctx, args); err != nil {
			skipped[i] = err
			// Mark as done
			progress.StatusBarCompletef(i, "Category skipped: %s", err.Error())
			continue
		}

		// Run categories concurrently
		categoriesWg.Add(1)
		go func(i int, category Category[Args]) {
			defer categoriesWg.Done()

			// Run all checks for this category concurrently
			var checksWg sync.WaitGroup
			var didErr atomic.Bool
			for _, check := range category.Checks {
				checksWg.Add(1)
				go func(check *Check[Args]) {
					defer updateChecksProgress()
					defer checksWg.Done()

					if err := check.IsEnabled(ctx, args); err != nil {
						progress.StatusBarUpdatef(i, "Check %s skipped: %s", check.Name, err.Error())
						return
					}

					// progress.Verbose never writes to output, so we just send check
					// progress to discard.
					var updateOutput strings.Builder
					if err := check.Update(ctx, std.NewFixedOutput(&updateOutput, true), args); err != nil {
						progress.StatusBarFailf(i, "Check %s failed: %s", check.Name, err.Error())
						check.cachedCheckOutput = updateOutput.String()
						didErr.Store(true)
					}
				}(check)
			}
			checksWg.Wait()

			// If error'd, status bar has already been set to failed with an error message
			if !didErr.Load() {
				progress.StatusBarCompletef(i, "Done!")
			}
		}(i, category)
	}
	categoriesWg.Wait()

	// Destroy progress and render a complete summary.
	// TODO we can probably refine and improe the summary a bit more.
	progress.Destroy()
	results := runAllCategoryChecksResult{
		categories: make(map[string]bool),
	}
	for i, category := range r.categories {
		results.all = append(results.all, i)
		idx := i + 1

		if _, ok := skipped[i]; ok {
			r.out.WriteSkippedf("%d. %s %s[SKIPPED. Reason: %s]%s", idx, category.Name,
				output.StyleBold, skipped[i], output.StyleReset)
			results.skipped = append(results.skipped, i)
			continue
		}

		// Report if this check is happy or not
		satisfied := category.IsSatisfied()
		results.categories[category.Name] = satisfied
		if satisfied {
			r.out.WriteSuccessf("%d. %s", idx, category.Name)
		} else {
			results.failed = append(results.failed, i)
			r.out.WriteFailuref("%d. %s", idx, category.Name)
			for _, c := range category.Checks {
				if c.cachedCheckErr != nil {
					r.out.WriteLine(output.Styledf(output.CombineStyles(output.StyleBold, output.StyleWarning),
						"%s: %s", c.Name, c.cachedCheckErr))
					// Render additional details
					if c.cachedCheckOutput != "" {
						r.out.Write(c.cachedCheckOutput)
					}
				}
			}
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

func (r *Runner[Args]) presentFailedCategoryWithOptions(ctx context.Context, categoryIdx int, category *Category[Args], args Args, results *runAllCategoryChecksResult) error {
	r.printCategoryHeaderAndDependencies(categoryIdx+1, category)
	fixableCategory := category.HasFixable()

	choices := map[int]string{}
	if fixableCategory {
		choices[1] = "You try fixing all of it for me."
		choices[2] = "I want to fix these manually"
		choices[3] = "Go back"
	} else {
		choices[1] = "I want to fix these manually"
		choices[2] = "Go back"
	}

	choice, err := getChoice(r.in, r.out, choices)
	if err != nil {
		return err
	}

	switch choice {
	case 1:
		if fixableCategory {
			r.out.ClearScreen()
			if !r.fixCategoryAutomatically(ctx, categoryIdx, category, args, results) {
				err = errors.Newf("%s: failed to fix category automatically", category.Name)
			}
		} else {
			err = r.fixCategoryManually(ctx, categoryIdx, category, args)
		}
	case 2:
		err = r.fixCategoryManually(ctx, categoryIdx, category, args)
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
		if dep.IsSatisfied() {
			r.out.WriteSuccessf("%d. %s", idx, dep.Name)
		} else {
			if dep.cachedCheckErr != nil {
				r.out.WriteFailuref("%d. %s: %s", idx, dep.Name, dep.cachedCheckErr)
			} else {
				r.out.WriteFailuref("%d. %s: %s", idx, dep.Name, "check failed")
			}
		}
	}
}

func (r *Runner[Args]) fixCategoryManually(ctx context.Context, categoryIdx int, category *Category[Args], args Args) error {
	for {
		toFix := []int{}

		for i, dep := range category.Checks {
			if dep.IsSatisfied() {
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
			idx, err = getNumberOutOf(r.in, r.out, toFix)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}

		check := category.Checks[idx]

		r.out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "%s", check.Name))
		r.out.Write("")

		if check.cachedCheckErr != nil {
			r.out.WriteLine(output.Styledf(output.StyleBold, "Check encountered the following error:\n\n%s%s\n", output.StyleReset, check.cachedCheckErr))
		}

		if check.Description == "" {
			return errors.Newf("No description available for manual fix - good luck!")
		}

		r.out.WriteLine(output.Styled(output.StyleBold, "How to fix:"))

		r.out.WriteMarkdown(check.Description)

		// Wait for user to finish
		r.out.Promptf("Hit 'Return' or 'Enter' when you are done.")
		waitForReturn(r.in)

		// Check statuses
		r.out.WriteLine(output.Styled(output.StylePending, "Determining status..."))
		if err := check.Update(ctx, r.out, args); err != nil {
			r.out.WriteWarningf("Check %q still not satisfied", check.Name)
			return err
		}

		// Print summary again
		r.printCategoryHeaderAndDependencies(categoryIdx, category)
	}

	return nil
}

func (r *Runner[Args]) fixCategoryAutomatically(ctx context.Context, categoryIdx int, category *Category[Args], args Args, results *runAllCategoryChecksResult) (ok bool) {
	// Best to be verbose when fixing, in case something goes wrong
	r.out.SetVerbose()
	defer r.out.UnsetVerbose()

	r.out.WriteLine(output.Styledf(output.StylePending, "Trying my hardest to fix %q automatically...", category.Name))

	// Make sure to call this with a final message before returning!
	complete := func(emoji string, style output.Style, fmtStr string, args ...any) {
		r.out.WriteLine(output.Linef(emoji, output.CombineStyles(style, output.StyleBold),
			"%d. %s - "+fmtStr, append([]any{categoryIdx, category.Name}, args...)...))
	}

	if err := category.CheckEnabled(ctx, args); err != nil {
		complete(output.EmojiQuestionMark, output.StyleGrey, "Skipped: %s", err.Error())
		return true
	}

	// If nothing in this category is fixable, we are done
	if !category.HasFixable() {
		complete(output.EmojiFailure, output.StyleFailure, "Cannot be fixed automatically.")
		return false
	}

	// Only run if dependents are fixed
	var unmetDependencies []string
	for _, d := range category.DependsOn {
		if met, exists := results.categories[d]; !exists {
			complete(output.EmojiFailure, output.StyleFailure, "Required check category %q not found", d)
			return false
		} else if !met {
			unmetDependencies = append(unmetDependencies, fmt.Sprintf("%q", d))
		}
	}
	if len(unmetDependencies) > 0 {
		complete(output.EmojiFailure, output.StyleFailure, "Required dependencies %s not met.", strings.Join(unmetDependencies, ", "))
		return false
	}

	// now go through the real dependencies
	for _, c := range category.Checks {
		// If category is fixed, we are good to go
		if c.IsSatisfied() {
			continue
		}

		// Skip
		if err := c.IsEnabled(ctx, args); err != nil {
			r.out.WriteLine(output.Linef(output.EmojiQuestionMark, output.CombineStyles(output.StyleGrey, output.StyleBold),
				"%q skipped: %s", c.Name, err.Error()))
			continue
		}

		// Otherwise, check if this is fixable at all
		if c.Fix == nil {
			r.out.WriteLine(output.Linef(output.EmojiShrug, output.CombineStyles(output.StyleWarning, output.StyleBold),
				"%q cannot be fixed automatically.", c.Name))
			continue
		}

		// Attempt fix. Get new args because things might have changed due to another
		// fix being run.
		r.out.VerboseLine(output.Linef(output.EmojiAsterisk, output.StylePending,
			"Fixing %q...", c.Name))
		if err := c.Fix(ctx, IO{
			Input:  r.in,
			Output: r.out,
		}, args); err != nil {
			r.out.WriteLine(output.Linef(output.EmojiWarning, output.CombineStyles(output.StyleFailure, output.StyleBold),
				"Failed to fix %q: %s", c.Name, err.Error()))
			continue
		}

		// Check is the fix worked
		if err := c.Update(ctx, r.out, args); err != nil {
			r.out.WriteLine(output.Styledf(output.CombineStyles(output.StyleWarning, output.StyleBold),
				"Check %q still failing: %s", c.Name, err.Error()))
		} else {
			r.out.WriteLine(output.Styledf(output.CombineStyles(output.StyleSuccess, output.StyleBold),
				"Check %q is satisfied now!", c.Name))
		}
	}

	ok = category.IsSatisfied()
	if ok {
		complete(output.EmojiSuccess, output.StyleSuccess, "Done!")
	} else {
		complete(output.EmojiFailure, output.StyleFailure, "Some checks are still not satisfied")
	}

	return
}
