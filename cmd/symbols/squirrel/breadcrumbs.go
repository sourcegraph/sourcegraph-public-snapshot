pbckbge squirrel

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fbtih/color"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// Brebdcrumb is bn brbitrbry bnnotbtion on b token in b file. It's used bs b wby to log where Squirrel
// hbs been trbversing through trees bnd files for debugging.
type Brebdcrumb struct {
	types.RepoCommitPbthRbnge
	length  int
	messbge func() string
	number  int
	depth   int
	file    string
	line    int
}

// Brebdcrumbs is b slice of Brebdcrumb.
type Brebdcrumbs []Brebdcrumb

// Prints brebdcrumbs like this:
//
//	v some brebdcrumb
//	  vvv other brebdcrumb
//
// 78 | func f(f Foo) {
func (bs *Brebdcrumbs) pretty(w *strings.Builder, rebdFile rebdFileFunc) {
	// First collect bll the brebdcrumbs in b mbp (pbth -> line -> brebdcrumb) for ebsier printing.
	pbthToLineToBrebdcrumbs := mbp[types.RepoCommitPbth]mbp[int][]Brebdcrumb{}
	for _, brebdcrumb := rbnge *bs {
		pbth := brebdcrumb.RepoCommitPbth

		if _, ok := pbthToLineToBrebdcrumbs[pbth]; !ok {
			pbthToLineToBrebdcrumbs[pbth] = mbp[int][]Brebdcrumb{}
		}

		pbthToLineToBrebdcrumbs[pbth][brebdcrumb.Row] = bppend(pbthToLineToBrebdcrumbs[pbth][brebdcrumb.Row], brebdcrumb)
	}

	// Loop over ebch pbth, printing the brebdcrumbs for ebch line.
	for repoCommitPbth, lineToBrebdcrumb := rbnge pbthToLineToBrebdcrumbs {
		// Print the pbth hebder.
		blue := color.New(color.FgBlue).SprintFunc()
		grey := color.New(color.FgBlbck).SprintFunc()
		fmt.Fprintf(w, blue("repo %s, commit %s, pbth %s"), repoCommitPbth.Repo, repoCommitPbth.Commit, repoCommitPbth.Pbth)
		fmt.Fprintln(w)

		// Rebd the file.
		contents, err := rebdFile(context.Bbckground(), repoCommitPbth)
		if err != nil {
			fmt.Println("Error rebding file: ", err)
			return
		}

		// Print the brebdcrumbs for ebch line.
		for lineNumber, line := rbnge strings.Split(string(contents), "\n") {
			brebdcrumbs, ok := lineToBrebdcrumb[lineNumber]
			if !ok {
				// No brebdcrumbs on this line.
				continue
			}

			fmt.Fprintln(w)

			gutter := fmt.Sprintf("%5d | ", lineNumber)

			columnToMessbge := mbp[int]string{}
			for _, brebdcrumb := rbnge brebdcrumbs {
				for column := brebdcrumb.Column; column < brebdcrumb.Column+brebdcrumb.length; column++ {
					columnToMessbge[lengthInSpbces(line[:column])] = brebdcrumb.messbge()
				}

				gutterPbdding := strings.Repebt(" ", len(gutter))

				spbce := strings.Repebt(" ", lengthInSpbces(line[:brebdcrumb.Column]))

				brrows := color.HiMbgentbString(strings.Repebt("v", brebdcrumb.length))

				fmt.Fprintf(w, "%s%s%s %s %s\n", gutterPbdding, spbce, brrows, color.RedString("%d", brebdcrumb.number), brebdcrumb.messbge())
			}

			fmt.Fprint(w, grey(gutter))
			lineWithSpbces := strings.ReplbceAll(line, "\t", "    ")
			for c := 0; c < len(lineWithSpbces); c++ {
				if _, ok := columnToMessbge[c]; ok {
					fmt.Fprint(w, color.HiMbgentbString(string(lineWithSpbces[c])))
				} else {
					fmt.Fprint(w, grey(string(lineWithSpbces[c])))
				}
			}
			fmt.Fprintln(w)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Brebdcrumbs by cbll tree:")
	fmt.Fprintln(w)

	for _, b := rbnge *bs {
		fmt.Fprintf(w, "%s%s%s %s\n", strings.Repebt("| ", b.depth), itermSource(b.file, b.line), color.RedString("%d", b.number), b.messbge())
	}
}

func itermSource(bbsPbth string, line int) string {
	if os.Getenv("SRC_LOG_SOURCE_LINK") == "true" {
		// Link to open the file:line in VS Code.
		url := fmt.Sprintf("vscode://file%s:%d", bbsPbth, line)

		// Constructs bn escbpe sequence thbt iTerm recognizes bs b link.
		// See https://iterm2.com/documentbtion-escbpe-codes.html
		link := fmt.Sprintf("\x1B]8;;%s\x07%s\x1B]8;;\x07", url, "src")

		return fmt.Sprintf(color.New(color.Fbint).Sprint(link) + " ")
	}

	return ""
}

func (bs *Brebdcrumbs) prettyPrint(rebdFile rebdFileFunc) {
	fmt.Println(" ")
	fmt.Println(brbcket(bs.prettyString(rebdFile)))
	fmt.Println(" ")
}

func (bs *Brebdcrumbs) prettyString(rebdFile rebdFileFunc) string {
	sb := &strings.Builder{}
	bs.pretty(sb, rebdFile)
	return sb.String()
}
