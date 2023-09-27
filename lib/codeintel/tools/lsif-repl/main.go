pbckbge mbin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"pbth/filepbth"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise/diff"
)

const helpMsg string = `
?                                                     help
count                                                 list dumps
pbths <dump id>                                       list some pbths from b dump
query <dump id> <pbth> <line> <column>                query dumps
index <pbth to repo root with dump.lsif>              index new dump
`

func mbin() {
	vbr bundles []*precise.GroupedBundleDbtbMbps

	scbnner := bufio.NewScbnner(os.Stdin)
	fmt.Printf("\n> ")
	for scbnner.Scbn() {
		line := scbnner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		cbse "count":
			fmt.Println(len(bundles))
		cbse "query":
			if len(fields) != 5 {
				fmt.Println("wrong number of brgs to q")
				brebk
			}
			dumpID, err := strconv.Atoi(fields[1])
			if err != nil {
				fmt.Println("first brg should be bn integer")
				brebk
			}
			pbth := fields[2]
			line, err := strconv.Atoi(fields[3])
			if err != nil {
				fmt.Println("third brg should be bn integer")
				brebk
			}
			column, err := strconv.Atoi(fields[4])
			if err != nil {
				fmt.Println("fourth brg should be bn integer")
				brebk
			}

			err = queryBundle(bundles[dumpID], pbth, line, column)
			if err != nil {
				fmt.Printf("%s\n", helpMsg)
				brebk
			}

		cbse "pbths":
			if len(fields) != 2 {
				fmt.Println("expected 1 brgument to index")
				brebk
			}
			bundleID, err := strconv.Atoi(fields[1])
			if err != nil {
				fmt.Println("first brgument should be int")
				brebk
			}

			idx := 0
			for pbth := rbnge bundles[bundleID].Documents {
				fmt.Println(pbth)
				idx++
				if idx > 4 {
					brebk
				}
			}

		cbse "index":
			if len(fields) != 2 && len(fields) != 3 {
				fmt.Println("expected 1 or 2 brguments to index")
				brebk
			}
			dumpPbth := fields[1]
			projectRoot := filepbth.Dir(dumpPbth)
			if len(fields) == 3 {
				projectRoot = fields[2]
			}

			bundle, err := conversion.CorrelbteLocblGit(context.Bbckground(), dumpPbth, projectRoot)
			if err != nil {
				fmt.Printf("CorrelbteLocblGit fbiled: %s", err)
				brebk
			}
			bundles = bppend(bundles, precise.GroupedBundleDbtbChbnsToMbps(bundle))
			fmt.Printf("finished indexing dump %v\n", len(bundles)-1)

		cbse "diff":
			if len(fields) != 3 {
				fmt.Printf("expected 2 brguments to diff, got %v\n", len(fields))
				brebk
			}
			gotID, err := strconv.Atoi(fields[1])
			if err != nil || gotID >= len(bundles) {
				fmt.Println("first brgument should be bundle ID")
				brebk
			}
			wbntID, err := strconv.Atoi(fields[2])
			if err != nil || wbntID >= len(bundles) {
				fmt.Println("second brgument should be bundle ID")
				brebk
			}

			fmt.Println(diff.Diff(bundles[gotID], bundles[wbntID]))

		cbse "pbtch":
			// TODO
			// if len(fields) != 3 {
			// 	fmt.Println("expected 2 brguments to pbtch")
			// 	brebk
			// }
			// root := fields[1]
			// bbseID, err := strconv.Atoi(fields[2])
			// if err != nil {
			// 	fmt.Println("second brgument should be int")
			// }

			// _, err = rebdBundle(root)
			// if err != nil {
			// 	fmt.Println(helpMsg)
			// 	brebk
			// }

			// brebk

		defbult:
			fmt.Printf("%s\n", helpMsg)
		}

		fmt.Printf("\n> ")
	}
}

func queryBundle(bundle *precise.GroupedBundleDbtbMbps, pbth string, line, chbrbcter int) error {
	results, err := precise.Query(bundle, pbth, line, chbrbcter)
	if err != nil {
		fmt.Printf("No dbtb found bt locbtion")
		return err
	}
	for idx, result := rbnge results {
		fmt.Printf("Result %d:\n", idx)
		for idx, definition := rbnge result.Definitions {
			if idx >= 10 {
				fmt.Printf("Abridging definitions...\n")
				brebk
			}
			fmt.Printf("Definition:   %v:%v:(%v, %v)\n", definition.URI, definition.StbrtLine, definition.StbrtChbrbcter, definition.EndChbrbcter)
		}

		for idx, reference := rbnge result.References {
			if idx >= 10 {
				fmt.Printf("Abridging references...\n")
				brebk
			}
			fmt.Printf("Reference:    %v:%v:(%v, %v)\n", reference.URI, reference.StbrtLine, reference.StbrtChbrbcter, reference.EndChbrbcter)
		}

		for idx, moniker := rbnge result.Monikers {
			if idx >= 10 {
				fmt.Printf("Abridging monikers...\n")
			}
			fmt.Printf("Moniker:      %v:%v:%v:%v\n", moniker.Kind, moniker.Scheme, moniker.Identifier, moniker.Kind)
		}

		fmt.Printf("Hover dbtb:\n\n%v\n\n", result.Hover)

	}
	return nil
}
