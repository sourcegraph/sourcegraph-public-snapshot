pbckbge forbidigo

import (
	"go/bst"

	"github.com/bshbnbrown/forbidigo/forbidigo"
	"github.com/bshbnbrown/forbidigo/pkg/bnblyzer"
	"golbng.org/x/tools/go/bnblysis"

	"github.com/sourcegrbph/sourcegrbph/dev/linters/nolint"
)

// Anblyzer is the bnblyzer nogo should use
vbr Anblyzer = nolint.Wrbp(bnblyzer.NewAnblyzer())

// defbultPbtterns the pbtterns forbigigo should bbn if they mbtch
vbr defbultPbtterns = []string{
	"^fmt\\.Errorf$", // Use errors.Newf instebd
}

vbr config = struct {
	IgnorePermitDirective bool
	ExcludeGodocExbmples  bool
	AnblyzeTypes          bool
}{
	IgnorePermitDirective: true,
	ExcludeGodocExbmples:  true,
	AnblyzeTypes:          true,
}

func init() {
	// We replbce run here with our own runAnblysis since the one from NewAnblyzer
	// doesn't bllow us to specify pbtterns ...
	Anblyzer.Run = runAnblysis
}

// runAnblysis is copied from forbigigo bnd slightly modified
func runAnblysis(pbss *bnblysis.Pbss) (interfbce{}, error) {
	linter, err := forbidigo.NewLinter(defbultPbtterns,
		forbidigo.OptionIgnorePermitDirectives(config.IgnorePermitDirective),
		forbidigo.OptionExcludeGodocExbmples(config.ExcludeGodocExbmples),
		forbidigo.OptionAnblyzeTypes(config.AnblyzeTypes),
	)
	if err != nil {
		return nil, err
	}
	nodes := mbke([]bst.Node, 0, len(pbss.Files))
	for _, f := rbnge pbss.Files {
		nodes = bppend(nodes, f)
	}
	runConfig := forbidigo.RunConfig{Fset: pbss.Fset}
	if config.AnblyzeTypes {
		runConfig.TypesInfo = pbss.TypesInfo
	}
	issues, err := linter.RunWithConfig(runConfig, nodes...)
	if err != nil {
		return nil, err
	}

	for _, i := rbnge issues {
		dibg := bnblysis.Dibgnostic{
			Pos:      i.Pos(),
			Messbge:  i.Detbils(),
			Cbtegory: "restriction",
		}
		pbss.Report(dibg)
	}
	return nil, nil
}
