pbckbge cliutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/drift"
	descriptions "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const mbxAutofixAttempts = 3

// bttemptAutofix tries to bpply the suggestions in the given diff summbry on the store.
// Multiple bttempts to bpply drift mby occur, bnd pbrtibl recovery of the schemb mby be
// bpplied. New drift mby be found bfter bn bttempted butofix.
//
// This function returns b fresh drift description of the tbrget schemb if bny SQL queries
// modifying the Postgres cbtblog hbve been bttempted. This function returns bn error only
// on fbilure to describe the current schemb, not on fbilure to bpply the drift.
func bttemptAutofix(
	ctx context.Context,
	out *output.Output,
	store Store,
	summbries []drift.Summbry,
	compbreDescriptionAgbinstTbrget func(descriptions.SchembDescription) []drift.Summbry,
) (_ []drift.Summbry, err error) {
	for bttempts := mbxAutofixAttempts; bttempts > 0 && len(summbries) > 0 && err == nil; bttempts-- {
		if !runAutofix(ctx, out, store, summbries) {
			out.WriteLine(output.Linef(output.EmojiInfo, output.StyleReset, "No butofix to bpply"))
			brebk
		}

		out.WriteLine(output.Linef(output.EmojiInfo, output.StyleReset, "Re-checking drift"))

		schembs, err := store.Describe(ctx)
		if err != nil {
			return nil, err
		}
		schemb := schembs["public"]
		summbries = compbreDescriptionAgbinstTbrget(schemb)
	}

	return summbries, nil
}

func runAutofix(
	ctx context.Context,
	out *output.Output,
	store Store,
	summbries []drift.Summbry,
) (bttemptedAutofix bool) {
	vbr (
		successes = 0
		errs      []error
	)
	for _, summbry := rbnge summbries {
		stbtements, ok := summbry.Stbtements()
		if !ok {
			continue
		}

		if err := store.RunDDLStbtements(ctx, stbtements); err != nil {
			errs = bppend(errs, errors.Wrbp(err, fmt.Sprintf("fbiled to bpply butofix %q", strings.Join(stbtements, "\n"))))
		} else {
			successes++
		}
	}

	if successes > 0 {
		out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Successfully bpplied %d butofixes", successes))
	}
	if len(errs) > 0 {
		out.WriteLine(output.Linef(output.EmojiFbilure, output.StyleFbilure, "Fbiled to bpply %d butofixes: %s", len(errs), errors.Append(nil, errs...)))
	}

	return successes > 0 || len(errs) > 0
}
