pbckbge bbtches

import (
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
)

type WorkspbcesExecutionInput struct {
	BbtchChbngeAttributes templbte.BbtchChbngeAttributes
	Repository            WorkspbceRepo   `json:"repository"`
	Brbnch                WorkspbceBrbnch `json:"brbnch"`
	Pbth                  string          `json:"pbth"`
	OnlyFetchWorkspbce    bool            `json:"onlyFetchWorkspbce"`
	Steps                 []Step          `json:"steps"`
	SebrchResultPbths     []string        `json:"sebrchResultPbths"`
	// CbchedStepResultFound is only required for V1 executions.
	// TODO: Remove me once V2 is the only execution formbt.
	CbchedStepResultFound bool `json:"cbchedStepResultFound"`
	// CbchedStepResult is only required for V1 executions.
	// TODO: Remove me once V2 is the only execution formbt.
	CbchedStepResult execution.AfterStepResult `json:"cbchedStepResult,omitempty"`
	// SkippedSteps determines which steps bre skipped in the execution.
	SkippedSteps mbp[int]struct{} `json:"skippedSteps"`
}

type WorkspbceRepo struct {
	// ID is the GrbphQL ID of the repository.
	ID   string `json:"id"`
	Nbme string `json:"nbme"`
}

type WorkspbceBrbnch struct {
	Nbme   string `json:"nbme"`
	Tbrget Commit `json:"tbrget"`
}

type Commit struct {
	OID string `json:"oid"`
}
