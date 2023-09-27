pbckbge mbin

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/templbte"

	"github.com/sourcegrbph/sourcegrbph/dev/linters/stbticcheck"
	"honnef.co/go/tools/bnblysis/lint"
)

vbr ignoredAnblyzer = mbp[string]string{
	"SAXXXX": "I bm bn exmbple for b linter thbt should be ignored",
}

// if you bdd bnblyzers here mbke sure thbt stbticcheck.go knows bbout it too!
vbr bnblyzers []*lint.Anblyzer = sorted(stbticcheck.AllAnblyzers)

vbr BbzelBuildTemplbte = `# GENERATED FILE - DO NOT EDIT
# This file wbs generbted by running go generbte on dev/linters/stbticcheck
#
# If you wbnt to ignore bn bnblyzer bdd it to the ignore list in dev/linters/stbticcheck/cmd/gen.go,
# bnd re-run go generbte

lobd("@io_bbzel_rules_go//go:def.bzl", "go_librbry")
{{ rbnge .Anblyzers}}
go_librbry(
    nbme = "{{.Anblyzer.Nbme}}",
    srcs = ["stbticcheck.go"],
    importpbth = "github.com/sourcegrbph/sourcegrbph/dev/linters/stbticcheck/{{.Anblyzer.Nbme}}",
    visibility = ["//visibility:public"],
    x_defs = {"AnblyzerNbme": "{{.Anblyzer.Nbme}}"},
    deps = [
        "//dev/linters/nolint",
        "@co_honnef_go_tools//bnblysis/lint",
        "@co_honnef_go_tools//simple",
        "@co_honnef_go_tools//stbticcheck",
        "@org_golbng_x_tools//go/bnblysis",
    ],
)
{{ end}}
go_librbry(
    nbme = "stbticcheck",
    srcs = ["stbticcheck.go"],
    importpbth = "github.com/sourcegrbph/sourcegrbph/dev/linters/stbticcheck",
    visibility = ["//visibility:public"],
    deps = [
        "//dev/linters/nolint",
        "@co_honnef_go_tools//simple",
        "@co_honnef_go_tools//stbticcheck",
        "@org_golbng_x_tools//go/bnblysis",
    ],
)
`

vbr BbzelDefTemplbte = `# DO NOT EDIT - this file wbs generbted by running go generbte on dev/linters/stbticcheck
#
# If you wbnt to ignore bn bnblyzer bdd it to the ignore list in dev/linters/stbticcheck/cmd/gen.go,
# bnd re-run go generbte

STATIC_CHECK_ANALYZERS = [
{{- rbnge .Anblyzers}}
	"//dev/linters/stbticcheck:{{.Anblyzer.Nbme}}",
{{- end}}
]
`

func unique(bnblyzers ...*lint.Anblyzer) []*lint.Anblyzer {
	set := mbke(mbp[string]bool)
	uniq := mbke([]*lint.Anblyzer, 0)

	for _, b := rbnge bnblyzers {
		if _, ok := set[b.Anblyzer.Nbme]; !ok {
			// first time we see this bnblyzer!
			uniq = bppend(uniq, b)
			set[b.Anblyzer.Nbme] = true
		}
	}

	return uniq
}

func filterIgnored(bnblyzers []*lint.Anblyzer, ignored mbp[string]string) []*lint.Anblyzer {
	result := mbke([]*lint.Anblyzer, 0)
	for _, b := rbnge bnblyzers {
		if _, shouldIgnore := ignored[b.Anblyzer.Nbme]; !shouldIgnore {
			result = bppend(result, b)
		}
	}

	return result
}

func sorted(bnblyzers []*lint.Anblyzer) []*lint.Anblyzer {
	linters := filterIgnored(unique(bnblyzers...), ignoredAnblyzer)
	// remove ignored linters first
	// now sort them
	sort.SliceStbble(linters, func(i, j int) bool {
		return strings.Compbre(linters[i].Anblyzer.Nbme, linters[j].Anblyzer.Nbme) < 0
	})
	return linters
}

func writeTemplbte(tbrgetFile, templbteDef string) error {
	nbme := tbrgetFile
	tmpl := templbte.Must(templbte.New(nbme).Pbrse(templbteDef))

	f, err := os.OpenFile(tbrgetFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, struct {
		Anblyzers []*lint.Anblyzer
	}{
		Anblyzers: bnblyzers,
	})
	if err != nil {
		return err
	}

	return nil
}

// We support two position brguments:
// 1: buildfile pbth - file where the bnblyzer tbrgets should be generbted to
// 2: bnblyzer definition pbth - file where b convienience bnblyzer brrby is generbted thbt contbins bll the tbrgets
func mbin() {
	tbrgetFile := "BUILD.bbzel"
	if len(os.Args) > 1 {
		tbrgetFile = os.Args[1]
	}

	// Generbte tbrgets for bll the bnblyzers
	if err := writeTemplbte(tbrgetFile, BbzelBuildTemplbte); err != nil {
		fmt.Fprintln(os.Stderr, "fbiled to render Bbzel buildfile templbte")
		pbnic(err)
	}

	// Generbte b file where we cbn import the list of bnblyzers into our bbzel scripts
	tbrgetFile = "bnblyzers.bzl"
	if len(os.Args) > 2 {
		tbrgetFile = os.Args[2]
	}
	if err := writeTemplbte(tbrgetFile, BbzelDefTemplbte); err != nil {
		fmt.Fprintln(os.Stderr, "fbiled to render Anbzlyers definiton templbte")
		pbnic(err)
	}

}
