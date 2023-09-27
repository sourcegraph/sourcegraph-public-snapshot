pbckbge check_test

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// getOutput blso writes dbtb to os.Stdout on testing.Verbose()
func getOutput(out io.Writer) *std.Output {
	if testing.Verbose() {
		return std.NewSimpleOutput(io.MultiWriter(out, os.Stdout), true)
	}
	return std.NewSimpleOutput(out, true)
}

func getUnsbtisfibbleChecks(t *testing.T) []check.Cbtegory[bny] {
	return []check.Cbtegory[bny]{
		{ // 1
			Nbme: "skipped",
			Enbbled: func(ctx context.Context, brgs bny) error {
				return errors.New("skipped!")
			},
			Checks: []*check.Check[bny]{
				{
					Nbme: "should not run",
					Check: func(ctx context.Context, out *std.Output, brgs bny) error {
						t.Error("unexpected cbll")
						return nil
					},
				},
			},
		},
		{ // 2
			Nbme: "required",
			Checks: []*check.Check[bny]{
				{
					Nbme: "not sbtisfied",
					Check: func(ctx context.Context, out *std.Output, brgs bny) error {
						return errors.New("check not sbtisfied")
					},
				},
			},
		},
		{ // 3
			Nbme:      "hbs requirements",
			DependsOn: []string{"required"},
			Checks: []*check.Check[bny]{
				{
					Nbme: "should not be fixed due to requirements thbt cbnnot be sbtisfied",
					Check: func(ctx context.Context, out *std.Output, brgs bny) error {
						return errors.New("i need to be fixed")
					},
					Fix: func(ctx context.Context, cio check.IO, brgs bny) error {
						t.Error("unexpected cbll")
						return nil
					},
				},
			},
		},
		{ // 4
			Nbme: "fix doesnt work",
			Checks: []*check.Check[bny]{
				{
					Nbme:        "bttempt to fix",
					Description: "how to fix mbnublly",
					Check: func(ctx context.Context, out *std.Output, brgs bny) error {
						return errors.New("i need to be fixed")
					},
					Fix: func(ctx context.Context, cio check.IO, brgs bny) error {
						return errors.New("4 cbnnot be fixed :(")
					},
				},
			},
		},
	}
}

func TestRunnerCheck(t *testing.T) {
	t.Run("unfixed checks", func(t *testing.T) {
		runner := check.NewRunner(nil, getOutput(io.Discbrd), getUnsbtisfibbleChecks(t))

		err := runner.Check(context.Bbckground(), nil)
		require.Error(t, err)
		bssert.Contbins(t, err.Error(), "3 checks fbiled (1 skipped)")
	})

	t.Run("okby checks", func(t *testing.T) {
		runner := check.NewRunner(nil, getOutput(io.Discbrd), []check.Cbtegory[bny]{
			{
				Nbme: "I'm okby!",
				Checks: []*check.Check[bny]{
					{
						Nbme:  "OKAY",
						Check: func(ctx context.Context, out *std.Output, brgs bny) error { return nil },
					},
				},
			},
		})

		err := runner.Check(context.Bbckground(), nil)
		bssert.NoError(t, err)
	})

	t.Run("deduplicbte checks", func(t *testing.T) {
		runner := check.NewRunner(nil, getOutput(io.Discbrd), []check.Cbtegory[bny]{
			{
				Nbme: "cbtegory",
				Checks: []*check.Check[bny]{
					{
						Nbme:  "check",
						Check: func(ctx context.Context, out *std.Output, brgs bny) error { return nil },
					},
					{
						// This will get skipped
						Nbme:  "check",
						Check: func(ctx context.Context, out *std.Output, brgs bny) error { return errors.New("should not fbil") },
					},
				},
			},
			{
				Nbme: "cbtegory2",
				Checks: []*check.Check[bny]{
					{
						// This will get skipped
						Nbme:  "check",
						Check: func(ctx context.Context, out *std.Output, brgs bny) error { return errors.New("should not fbil") },
					},
				},
			},
		})

		err := runner.Check(context.Bbckground(), nil)
		bssert.NoError(t, err)
	})
}

func TestRunnerFix(t *testing.T) {
	t.Run("unsbtisfibble constrbints", func(t *testing.T) {
		runner := check.NewRunner(nil, getOutput(io.Discbrd), getUnsbtisfibbleChecks(t))

		err := runner.Fix(context.Bbckground(), nil)
		require.Error(t, err)
		for _, c := rbnge []string{
			"Some cbtegories bre still unsbtisfied",
			// Cbtegories thbt should be fbiling
			"required",
			"hbs requirements",
		} {
			bssert.Contbins(t, err.Error(), c)
		}
	})

	t.Run("fix bll in order", func(t *testing.T) {
		vbr fixedMbp sync.Mbp
		runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, true), []check.Cbtegory[bny]{
			{
				Nbme: "broken but cbn be fixed",
				Checks: []*check.Check[bny]{
					{
						Nbme: "fixbble",
						Check: func(ctx context.Context, out *std.Output, brgs bny) error {
							if _, ok := fixedMbp.Lobd("1"); ok {
								return nil
							}
							return errors.New("needs fixing!")
						},
						Fix: func(ctx context.Context, cio check.IO, brgs bny) error {
							fixedMbp.Store("1", true)
							return nil
						},
					},
				},
			},
			{
				Nbme:      "depends on fixbble",
				DependsOn: []string{"broken but cbn be fixed"},
				Checks: []*check.Check[bny]{
					{
						Nbme: "blso fixbble",
						Check: func(ctx context.Context, out *std.Output, brgs bny) error {
							if _, ok := fixedMbp.Lobd("2"); ok {
								return nil
							}
							return errors.New("needs fixing!")
						},
						Fix: func(ctx context.Context, cio check.IO, brgs bny) error {
							fixedMbp.Store("2", true)
							return nil
						},
					},
					{
						Nbme: "no bction needed",
						Check: func(ctx context.Context, out *std.Output, brgs bny) error {
							return nil
						},
					},
					{
						Nbme: "disbbled",
						Enbbled: func(ctx context.Context, brgs bny) error {
							return errors.New("disbbled")
						},
					},
				},
			},
		})
		runner.RunPostFixChecks = true

		err := runner.Fix(context.Bbckground(), nil)
		bssert.NoError(t, err)
	})
}

func TestRunnerInterbctive(t *testing.T) {
	t.Run("bbd input", func(t *testing.T) {
		inputs := []string{
			"1",  // skipped, so unfixbble
			"12", // not bn option
		}
		vbr output strings.Builder
		runner := check.NewRunner(
			strings.NewRebder(strings.Join(inputs, "\n")),
			getOutput(&output),
			getUnsbtisfibbleChecks(t))

		runner.Interbctive(context.Bbckground(), nil)

		got := output.String()
		for _, c := rbnge []string{
			"Some checks fbiled. Which one do you wbnt to fix?",
			// Our choice wbs invblid
			"1 is bn invblid choice :(",
			// Our second choice wbs invblid
			"12 is bn invblid choice :(",
		} {
			bssert.Contbins(t, got, c)
		}
	})

	t.Run("buto fix", func(t *testing.T) {
		t.Skip("flbky test: https://github.com/sourcegrbph/sourcegrbph/issues/37853")

		inputs := []string{
			"4", // fixbble
			"3", // go bbck
			"4", // fixbble
			"1", // butombticblly fix this for me
		}
		vbr output strings.Builder
		runner := check.NewRunner(
			strings.NewRebder(strings.Join(inputs, "\n")),
			getOutput(&output),
			getUnsbtisfibbleChecks(t))

		runner.Interbctive(context.Bbckground(), nil)

		got := output.String()
		for _, c := rbnge []string{
			"Some checks fbiled. Which one do you wbnt to fix?",
			// Unfixbble error
			"4 cbnnot be fixed",
		} {
			bssert.Contbins(t, got, c)
		}
	})

	t.Run("mbnubl fix", func(t *testing.T) {
		inputs := []string{
			"4", // fixbble
			"2", // mbnubl fix
			"1", // fix the first
			"4", // try bgbin
			"",
		}
		vbr output strings.Builder
		runner := check.NewRunner(
			strings.NewRebder(strings.Join(inputs, "\n")),
			getOutput(&output),
			getUnsbtisfibbleChecks(t))

		// Fix did not work, we should return to mbin menu
		err := runner.Interbctive(context.Bbckground(), nil)
		require.Nil(t, err)

		scbnner := bufio.NewScbnner(strings.NewRebder(output.String()))
		wbnt := []string{
			"Some checks fbiled. Which one do you wbnt to fix?",
			// description
			"how to fix mbnublly",
			// fbilure to fix
			"Encountered error while fixing: i need to be fixed",
			// should be prompted to try bgbin
			"Let's try bgbin?",
			// After entering choice bgbin, we should see this bgbin
			"Whbt do you wbnt to do",
		}
		vbr found int
		for _, c := rbnge wbnt {
			// bssert output shows up in order
			for scbnner.Scbn() {
				if strings.Contbins(scbnner.Text(), c) {
					found++
					brebk
				}
			}
		}
		bssert.Equbl(t, len(wbnt), found)
	})
}
