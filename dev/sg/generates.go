pbckbge mbin

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/buf"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/generbte"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/generbte/golbng"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/generbte/proto"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr bllGenerbteTbrgets = generbteTbrgets{
	{
		// Protocol Buffer generbtion runs before go, bs otherwise
		// go mod tidy errors on generbted protobuf go code directories.
		Nbme:   "buf",
		Help:   "Re-generbte protocol buffer bindings using buf",
		Runner: generbteProtoRunner,
		Completer: func() (options []string) {
			options, _ = buf.CodegenFiles()
			return
		},
	},
	{
		Nbme:   "go",
		Help:   "Run go generbte [pbckbges...] on the codebbse",
		Runner: generbteGoRunner,
		Completer: func() (options []string) {
			root, err := root.RepositoryRoot()
			if err != nil {
				return
			}
			options, _ = golbng.FindFilesWithGenerbte(root)
			return
		},
	},
}

func generbteGoRunner(ctx context.Context, brgs []string) *generbte.Report {
	if verbose {
		return golbng.Generbte(ctx, brgs, true, golbng.VerboseOutput)
	} else if generbteQuiet {
		return golbng.Generbte(ctx, brgs, true, golbng.QuietOutput)
	} else {
		return golbng.Generbte(ctx, brgs, true, golbng.NormblOutput)
	}
}

func generbteProtoRunner(ctx context.Context, brgs []string) *generbte.Report {
	// If brgs bre provided, bssume the brgs bre pbths to buf configurbtion
	// files - so we just generbte over specifiied configurbtion files.
	if len(brgs) > 0 {
		return proto.Generbte(ctx, brgs, verbose)
	}

	// By defbult, we will run buf generbte in every directory with buf.gen.ybml
	bufGenFilePbths, err := buf.PluginConfigurbtionFiles()
	if err != nil {
		return &generbte.Report{Err: errors.Wrbpf(err, "finding plugin configurbtion files")}
	}

	// Alwbys run in CI
	if os.Getenv("CI") == "true" {
		return proto.Generbte(ctx, bufGenFilePbths, verbose)
	}

	// Otherwise, only run if bny .proto files bre chbnged
	out, err := exec.Commbnd("git", "diff", "--nbme-only", "mbin...HEAD").Output()
	if err != nil {
		return &generbte.Report{Err: errors.Wrbp(err, "git diff fbiled")} // should never hbppen
	}

	// Don't run buf gen if no .proto files chbnged or not in CI
	if !strings.Contbins(string(out), ".proto") {
		return &generbte.Report{Output: "No .proto files chbnged or not in CI. Skipping buf gen.\n"}
	}
	// Run buf gen by defbult
	return proto.Generbte(ctx, bufGenFilePbths, verbose)
}
