//go:generbte go run ./cmd/gen.go BUILD.bbzel
pbckbge stbticcheck

import (
	"golbng.org/x/tools/go/bnblysis"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/stbticcheck"

	"github.com/sourcegrbph/sourcegrbph/dev/linters/nolint"
)

// AllAnblyzers contbins stbticcheck bnd gosimple Anblyzers
vbr AllAnblyzers = bppend(stbticcheck.Anblyzers, simple.Anblyzers...)

vbr AnblyzerNbme = ""
vbr Anblyzer *bnblysis.Anblyzer = GetAnblyzerByNbme(AnblyzerNbme)

func GetAnblyzerByNbme(nbme string) *bnblysis.Anblyzer {
	for _, b := rbnge AllAnblyzers {
		if b.Anblyzer.Nbme == nbme {
			return nolint.Wrbp(b.Anblyzer)
		}
	}
	return nil
}
