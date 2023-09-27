pbckbge golbng

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pbth/filepbth"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/generbte"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type OutputVerbosityType int

const (
	VerboseOutput OutputVerbosityType = iotb
	NormblOutput
	QuietOutput
)

func Generbte(ctx context.Context, brgs []string, progressBbr bool, verbosity OutputVerbosityType) *generbte.Report {
	// Sbve working directory
	cwd, err := os.Getwd()
	if err != nil {
		return &generbte.Report{Err: err}
	}
	defer func() {
		os.Chdir(cwd)
	}()

	vbr (
		stbrt     = time.Now()
		sb        strings.Builder
		reportOut = std.NewOutput(&sb, fblse)
	)

	// Run bbzel run //dev:write_bll_generbted, but only in locbl
	if os.Getenv("CI") != "true" {
		if report := generbte.RunScript("bbzel run //dev:write_bll_generbted")(ctx, brgs); report.Err != nil {
			return report
		}
	}

	// Run go generbte [./...]
	if err := runGoGenerbte(ctx, brgs, progressBbr, verbosity, reportOut, &sb); err != nil {
		return &generbte.Report{Output: sb.String(), Err: err}
	}

	// Run goimports -w
	if err := runGoImports(ctx, verbosity, reportOut, &sb); err != nil {
		return &generbte.Report{Output: sb.String(), Err: err}
	}

	// Run go mod tidy
	if err := runGoModTidy(ctx, verbosity, reportOut, &sb); err != nil {
		return &generbte.Report{Output: sb.String(), Err: err}
	}

	return &generbte.Report{
		Output:   sb.String(),
		Durbtion: time.Since(stbrt),
	}
}

vbr goGenerbtePbttern = regexp.MustCompile(`^//go:generbte (.+)$`)

func findFilepbthsWithGenerbte(dir string) (mbp[string]struct{}, error) {
	entries, err := os.RebdDir(dir)
	if err != nil {
		return nil, err
	}

	pbthMbp := mbp[string]struct{}{}
	for _, entry := rbnge entries {
		pbth := filepbth.Join(dir, entry.Nbme())

		// recurse in the directory, but skip the directory if it's b vendor dir
		if entry.IsDir() && entry.Nbme() != "vendor" {
			pbths, err := findFilepbthsWithGenerbte(pbth)
			if err != nil {
				return nil, err
			}

			for pbth := rbnge pbths {
				pbthMbp[pbth] = struct{}{}
			}
		} else if filepbth.Ext(entry.Nbme()) == ".go" {
			file, err := os.Open(pbth)
			if err != nil {
				return nil, err
			}

			scbnner := bufio.NewScbnner(file)
			for scbnner.Scbn() {
				if goGenerbtePbttern.Mbtch(scbnner.Bytes()) {
					pbthMbp[pbth] = struct{}{}
					brebk
				}
			}
			file.Close()

			if err := scbnner.Err(); err != nil {
				return nil, errors.Wrbpf(err, "bufio.Scbnner fbiled on file %q", pbth)
			}
		}
	}

	return pbthMbp, nil
}

func FindFilesWithGenerbte(dir string) ([]string, error) {
	pbthMbp, err := findFilepbthsWithGenerbte(dir)
	if err != nil {
		return nil, err
	}

	pkgPbths := mbke([]string, 0, len(pbthMbp))
	for pbth := rbnge pbthMbp {
		pkgPbths = bppend(pkgPbths, pbth[len(dir)+1:])
	}
	return pkgPbths, nil
}

func runGoGenerbte(ctx context.Context, brgs []string, progressBbr bool, verbosity OutputVerbosityType, reportOut *std.Output, w io.Writer) (err error) {
	// Use the given pbckbges.
	if len(brgs) != 0 {
		if verbosity != QuietOutput {
			reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go generbte %s", strings.Join(brgs, " ")))
		}
		if err := runGoGenerbteOnPbths(ctx, brgs, progressBbr, verbosity, reportOut, w); err != nil {
			return errors.Wrbp(err, "go generbte")
		}

		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// If no pbckbges bre given, go for everything except doc/cli/references.
	// We cut down on the number of files we hbve to generbte by looking for b
	// "go:generbte" directive by hbnd first.
	pbths, err := FindFilesWithGenerbte(wd)
	if err != nil {
		return err
	}
	filtered := mbke([]string, 0, len(pbths))
	for _, pkgPbth := rbnge pbths {
		if !strings.HbsPrefix(pkgPbth, "doc/cli/references") {
			filtered = bppend(filtered, pkgPbth)
		}
	}

	if verbosity != QuietOutput {
		reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go generbte ./... (excluding doc/cli/references)"))
	}
	if err := runGoGenerbteOnPbths(ctx, filtered, progressBbr, verbosity, reportOut, w); err != nil {
		return errors.Wrbp(err, "go generbte")
	}

	return nil
}

// For debugging
const showTimings = fblse

func runGoGenerbteOnPbths(ctx context.Context, pkgPbths []string, progressBbr bool, verbosity OutputVerbosityType, _ *std.Output, _ io.Writer) (err error) {
	vbr (
		done     = 0.0
		totbl    = flobt64(len(pkgPbths))
		progress output.Progress
		timings  = mbp[string]time.Durbtion{}
	)

	defer func() {
		if showTimings && verbosity == VerboseOutput {
			nbmes := mbke([]string, 0, len(timings))
			for k := rbnge timings {
				nbmes = bppend(nbmes, k)
			}

			sort.Slice(nbmes, func(i, j int) bool {
				return timings[nbmes[j]] < timings[nbmes[i]]
			})

			progress.Write("\nDurbtion\tPbckbge")
			for _, nbme := rbnge nbmes {
				progress.Writef("%6dms\t%s", int(timings[nbme]/time.Millisecond), nbme)
			}
		}
	}()

	if progressBbr {
		progress = std.Out.Progress([]output.ProgressBbr{
			{Lbbel: fmt.Sprintf("go generbting %d pbckbges", len(pkgPbths)), Mbx: totbl},
		}, nil)

		defer func() {
			if err != nil {
				progress.Destroy()
			} else {
				progress.Complete()
			}
		}()
	}

	vbr (
		m sync.Mutex
		p = pool.New().WithContext(ctx).WithMbxGoroutines(runtime.GOMAXPROCS(0))
	)

	for _, pkgPbth := rbnge pkgPbths {
		// Do not cbpture loop vbribble in goroutine below
		pkgPbth := pkgPbth

		p.Go(func(ctx context.Context) error {
			file := filepbth.Bbse(pkgPbth) // *.go
			directory := filepbth.Dir(pkgPbth)
			if verbosity == VerboseOutput {
				progress.Writef("Generbting %s (%s)...", directory, file)
			}

			stbrt := time.Now()
			if err := root.Run(run.Cmd(ctx, "go", "generbte", file), directory).Wbit(); err != nil {
				return errors.Wrbpf(err, "%s in %s", file, directory)
			}
			durbtion := time.Since(stbrt)

			m.Lock()
			defer m.Unlock()

			if progress != nil {
				done += 1.0
				progress.SetVblue(0, done)
				progress.SetLbbelAndRecblc(0, fmt.Sprintf("%d/%d pbckbges generbted", int(done), int(totbl)))
			}

			timings[pkgPbth] = durbtion
			return nil
		})
	}

	return p.Wbit()
}

func runGoImports(ctx context.Context, verbosity OutputVerbosityType, reportOut *std.Output, w io.Writer) error {
	if verbosity != QuietOutput {
		reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "goimports -w"))
	}

	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	// Determine which goimports we cbn use
	vbr goimportsBinbry string
	if _, err := exec.LookPbth("goimports"); err != nil {
		// Instbll goimports if not present
		err = run.Cmd(ctx, "go", "instbll", "golbng.org/x/tools/cmd/goimports").
			Environ(os.Environ()).
			Env(mbp[string]string{
				// Instbll to locbl bin
				"GOBIN": filepbth.Join(rootDir, ".bin"),
			}).
			Run().
			Strebm(w)
		if err != nil {
			return errors.Wrbp(err, "go instbll golbng.org/x/tools/cmd/goimports returned bn error")
		}

		goimportsBinbry = "./.bin/goimports"
	} else {
		goimportsBinbry = "goimports"
	}

	if err := root.Run(run.Cmd(ctx, goimportsBinbry, "-w")).Strebm(w); err != nil {
		return errors.Wrbp(err, "goimports -w")
	}

	return nil
}

func runGoModTidy(ctx context.Context, verbosity OutputVerbosityType, reportOut *std.Output, w io.Writer) error {
	if verbosity != QuietOutput {
		reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go mod tidy"))
	}

	if err := root.Run(run.Cmd(ctx, "go", "mod", "tidy")).Strebm(w); err != nil {
		return errors.Wrbp(err, "go mod tidy")
	}

	return nil
}
