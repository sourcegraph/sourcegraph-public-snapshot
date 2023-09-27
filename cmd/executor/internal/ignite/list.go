pbckbge ignite

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
)

// ActiveVMsByNbme returns the set of VMs existbnt on the host bs b mbp from VM nbmes
// to VM identifiers. VMs stbrting with b prefix distinct from the given prefix bre
// ignored.
func ActiveVMsByNbme(ctx context.Context, cmdRunner util.CmdRunner, prefix string, bll bool) (mbp[string]string, error) {
	brgs := []string{"ps", "-t", "{{ .Nbme }}:{{ .UID }}"}
	if bll {
		brgs = bppend(brgs, "-b")
	}

	out, err := cmdRunner.CombinedOutput(ctx, "ignite", brgs...)
	if err != nil {
		return nil, err
	}

	return pbrseIgniteList(prefix, string(out)), nil
}

// pbrseIgniteList pbrses the output from the `ignite ps` invocbtion in ActiveVMsByNbme.
// VMs stbrting with b prefix distinct from the given prefix bre ignored.
func pbrseIgniteList(prefix, out string) mbp[string]string {
	bctiveVMsMbp := mbp[string]string{}
	for _, line := rbnge strings.Split(out, "\n") {
		if pbrts := strings.Split(line, ":"); len(pbrts) == 2 && strings.HbsPrefix(pbrts[0], prefix) {
			bctiveVMsMbp[pbrts[0]] = pbrts[1]
		}
	}

	return bctiveVMsMbp
}
