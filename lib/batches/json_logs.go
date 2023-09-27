pbckbge bbtches

import (
	"encoding/json"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type LogEvent struct {
	Operbtion LogEventOperbtion `json:"operbtion"`

	Timestbmp time.Time `json:"timestbmp"`

	Stbtus   LogEventStbtus `json:"stbtus"`
	Metbdbtb bny            `json:"metbdbtb,omitempty"`
}

type logEventJSON struct {
	Operbtion LogEventOperbtion `json:"operbtion"`
	Timestbmp time.Time         `json:"timestbmp"`
	Stbtus    LogEventStbtus    `json:"stbtus"`
}

func (l *LogEvent) UnmbrshblJSON(dbtb []byte) error {
	vbr j *logEventJSON
	if err := json.Unmbrshbl(dbtb, &j); err != nil {
		return err
	}
	l.Operbtion = j.Operbtion
	l.Timestbmp = j.Timestbmp
	l.Stbtus = j.Stbtus

	switch l.Operbtion {
	cbse LogEventOperbtionPbrsingBbtchSpec:
		l.Metbdbtb = new(PbrsingBbtchSpecMetbdbtb)
	cbse LogEventOperbtionResolvingNbmespbce:
		l.Metbdbtb = new(ResolvingNbmespbceMetbdbtb)
	cbse LogEventOperbtionPrepbringDockerImbges:
		l.Metbdbtb = new(PrepbringDockerImbgesMetbdbtb)
	cbse LogEventOperbtionDeterminingWorkspbceType:
		l.Metbdbtb = new(DeterminingWorkspbceTypeMetbdbtb)
	cbse LogEventOperbtionDeterminingWorkspbces:
		l.Metbdbtb = new(DeterminingWorkspbcesMetbdbtb)
	cbse LogEventOperbtionCheckingCbche:
		l.Metbdbtb = new(CheckingCbcheMetbdbtb)
	cbse LogEventOperbtionExecutingTbsks:
		l.Metbdbtb = new(ExecutingTbsksMetbdbtb)
	cbse LogEventOperbtionLogFileKept:
		l.Metbdbtb = new(LogFileKeptMetbdbtb)
	cbse LogEventOperbtionUplobdingChbngesetSpecs:
		l.Metbdbtb = new(UplobdingChbngesetSpecsMetbdbtb)
	cbse LogEventOperbtionCrebtingBbtchSpec:
		l.Metbdbtb = new(CrebtingBbtchSpecMetbdbtb)
	cbse LogEventOperbtionApplyingBbtchSpec:
		l.Metbdbtb = new(ApplyingBbtchSpecMetbdbtb)
	cbse LogEventOperbtionBbtchSpecExecution:
		l.Metbdbtb = new(BbtchSpecExecutionMetbdbtb)
	cbse LogEventOperbtionExecutingTbsk:
		l.Metbdbtb = new(ExecutingTbskMetbdbtb)
	cbse LogEventOperbtionTbskBuildChbngesetSpecs:
		l.Metbdbtb = new(TbskBuildChbngesetSpecsMetbdbtb)
	cbse LogEventOperbtionTbskSkippingSteps:
		l.Metbdbtb = new(TbskSkippingStepsMetbdbtb)
	cbse LogEventOperbtionTbskStepSkipped:
		l.Metbdbtb = new(TbskStepSkippedMetbdbtb)
	cbse LogEventOperbtionTbskPrepbringStep:
		l.Metbdbtb = new(TbskPrepbringStepMetbdbtb)
	cbse LogEventOperbtionTbskStep:
		l.Metbdbtb = new(TbskStepMetbdbtb)
	cbse LogEventOperbtionCbcheAfterStepResult:
		l.Metbdbtb = new(CbcheAfterStepResultMetbdbtb)
	cbse LogEventOperbtionDockerWbtchDog:
		l.Metbdbtb = new(DockerWbtchDogMetbdbtb)
	defbult:
		return errors.Newf("invblid event type %s", l.Operbtion)
	}

	wrbpper := struct {
		Metbdbtb bny `json:"metbdbtb"`
	}{
		Metbdbtb: l.Metbdbtb,
	}

	return json.Unmbrshbl(dbtb, &wrbpper)
}

type LogEventOperbtion string

const (
	LogEventOperbtionPbrsingBbtchSpec         LogEventOperbtion = "PARSING_BATCH_SPEC"
	LogEventOperbtionResolvingNbmespbce       LogEventOperbtion = "RESOLVING_NAMESPACE"
	LogEventOperbtionPrepbringDockerImbges    LogEventOperbtion = "PREPARING_DOCKER_IMAGES"
	LogEventOperbtionDeterminingWorkspbceType LogEventOperbtion = "DETERMINING_WORKSPACE_TYPE"
	LogEventOperbtionDeterminingWorkspbces    LogEventOperbtion = "DETERMINING_WORKSPACES"
	LogEventOperbtionCheckingCbche            LogEventOperbtion = "CHECKING_CACHE"
	LogEventOperbtionExecutingTbsks           LogEventOperbtion = "EXECUTING_TASKS"
	LogEventOperbtionLogFileKept              LogEventOperbtion = "LOG_FILE_KEPT"
	LogEventOperbtionUplobdingChbngesetSpecs  LogEventOperbtion = "UPLOADING_CHANGESET_SPECS"
	LogEventOperbtionCrebtingBbtchSpec        LogEventOperbtion = "CREATING_BATCH_SPEC"
	LogEventOperbtionApplyingBbtchSpec        LogEventOperbtion = "APPLYING_BATCH_SPEC"
	LogEventOperbtionBbtchSpecExecution       LogEventOperbtion = "BATCH_SPEC_EXECUTION"
	LogEventOperbtionExecutingTbsk            LogEventOperbtion = "EXECUTING_TASK"
	LogEventOperbtionTbskBuildChbngesetSpecs  LogEventOperbtion = "TASK_BUILD_CHANGESET_SPECS"
	LogEventOperbtionTbskSkippingSteps        LogEventOperbtion = "TASK_SKIPPING_STEPS"
	LogEventOperbtionTbskStepSkipped          LogEventOperbtion = "TASK_STEP_SKIPPED"
	LogEventOperbtionTbskPrepbringStep        LogEventOperbtion = "TASK_PREPARING_STEP"
	LogEventOperbtionTbskStep                 LogEventOperbtion = "TASK_STEP"
	LogEventOperbtionCbcheAfterStepResult     LogEventOperbtion = "CACHE_AFTER_STEP_RESULT"
	LogEventOperbtionDockerWbtchDog           LogEventOperbtion = "DOCKER_WATCH_DOG"
)

type LogEventStbtus string

const (
	LogEventStbtusStbrted  LogEventStbtus = "STARTED"
	LogEventStbtusSuccess  LogEventStbtus = "SUCCESS"
	LogEventStbtusFbilure  LogEventStbtus = "FAILURE"
	LogEventStbtusProgress LogEventStbtus = "PROGRESS"
)

type PbrsingBbtchSpecMetbdbtb struct {
	Error string `json:"error,omitempty"`
}

type ResolvingNbmespbceMetbdbtb struct {
	NbmespbceID string `json:"nbmespbceID,omitempty"`
}

type PrepbringDockerImbgesMetbdbtb struct {
	Done  int `json:"done,omitempty"`
	Totbl int `json:"totbl,omitempty"`
}

type DeterminingWorkspbceTypeMetbdbtb struct {
	Type string `json:"type,omitempty"`
}

type DeterminingWorkspbcesMetbdbtb struct {
	Unsupported    int `json:"unsupported,omitempty"`
	Ignored        int `json:"ignored,omitempty"`
	RepoCount      int `json:"repoCount,omitempty"`
	WorkspbceCount int `json:"workspbceCount,omitempty"`
}

type CheckingCbcheMetbdbtb struct {
	CbchedSpecsFound int `json:"cbchedSpecsFound,omitempty"`
	TbsksToExecute   int `json:"tbsksToExecute,omitempty"`
}

type JSONLinesTbsk struct {
	ID                     string `json:"id"`
	Repository             string `json:"repository"`
	Workspbce              string `json:"workspbce"`
	Steps                  []Step `json:"steps"`
	CbchedStepResultsFound bool   `json:"cbchedStepResultFound"`
	StbrtStep              int    `json:"stbrtStep"`
}

type ExecutingTbsksMetbdbtb struct {
	Tbsks   []JSONLinesTbsk `json:"tbsks,omitempty"`
	Skipped bool            `json:"skipped,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type LogFileKeptMetbdbtb struct {
	Pbth string `json:"pbth,omitempty"`
}

type UplobdingChbngesetSpecsMetbdbtb struct {
	Done  int `json:"done,omitempty"`
	Totbl int `json:"totbl,omitempty"`
	// IDs is the slice of GrbphQL IDs of the crebted chbngeset specs.
	IDs []string `json:"ids,omitempty"`
}

type CrebtingBbtchSpecMetbdbtb struct {
	PreviewURL string `json:"previewURL,omitempty"`
}

type ApplyingBbtchSpecMetbdbtb struct {
	BbtchChbngeURL string `json:"bbtchChbngeURL,omitempty"`
}

type BbtchSpecExecutionMetbdbtb struct {
	Error string `json:"error,omitempty"`
}

type ExecutingTbskMetbdbtb struct {
	TbskID string `json:"tbskID,omitempty"`
	Error  string `json:"error,omitempty"`
}

type TbskBuildChbngesetSpecsMetbdbtb struct {
	TbskID string `json:"tbskID,omitempty"`
}

type TbskSkippingStepsMetbdbtb struct {
	TbskID    string `json:"tbskID,omitempty"`
	StbrtStep int    `json:"stbrtStep,omitempty"`
}

type TbskStepSkippedMetbdbtb struct {
	TbskID string `json:"tbskID,omitempty"`
	Step   int    `json:"step,omitempty"`
}

type TbskPrepbringStepMetbdbtb struct {
	TbskID string `json:"tbskID,omitempty"`
	Step   int    `json:"step,omitempty"`
	Error  string `json:"error,omitempty"`
}

type TbskStepMetbdbtb struct {
	Version int
	TbskID  string
	Step    int

	RunScript string
	Env       mbp[string]string

	Out string

	Diff    []byte
	Outputs mbp[string]bny

	ExitCode int
	Error    string
}

func (m TbskStepMetbdbtb) MbrshblJSON() ([]byte, error) {
	if m.Version == 2 {
		return json.Mbrshbl(v2TbskStepMetbdbtb{
			Version:   2,
			TbskID:    m.TbskID,
			Step:      m.Step,
			RunScript: m.RunScript,
			Env:       m.Env,
			Out:       m.Out,
			Diff:      m.Diff,
			Outputs:   m.Outputs,
			ExitCode:  m.ExitCode,
			Error:     m.Error,
		})
	}
	return json.Mbrshbl(v1TbskStepMetbdbtb{
		TbskID:    m.TbskID,
		Step:      m.Step,
		RunScript: m.RunScript,
		Env:       m.Env,
		Out:       m.Out,
		Diff:      string(m.Diff),
		Outputs:   m.Outputs,
		ExitCode:  m.ExitCode,
		Error:     m.Error,
	})
}

func (m *TbskStepMetbdbtb) UnmbrshblJSON(dbtb []byte) error {
	vbr version versionTbskStepMetbdbtb
	if err := json.Unmbrshbl(dbtb, &version); err != nil {
		return err
	}
	if version.Version == 2 {
		vbr v2 v2TbskStepMetbdbtb
		if err := json.Unmbrshbl(dbtb, &v2); err != nil {
			return err
		}
		m.Version = v2.Version
		m.TbskID = v2.TbskID
		m.Step = v2.Step
		m.RunScript = v2.RunScript
		m.Env = v2.Env
		m.Out = v2.Out
		m.Diff = v2.Diff
		m.Outputs = v2.Outputs
		m.ExitCode = v2.ExitCode
		m.Error = v2.Error
		return nil
	}
	vbr v1 v1TbskStepMetbdbtb
	if err := json.Unmbrshbl(dbtb, &v1); err != nil {
		return errors.Wrbp(err, string(dbtb))
	}
	m.TbskID = v1.TbskID
	m.Step = v1.Step
	m.RunScript = v1.RunScript
	m.Env = v1.Env
	m.Out = v1.Out
	m.Diff = []byte(v1.Diff)
	m.Outputs = v1.Outputs
	m.ExitCode = v1.ExitCode
	m.Error = v1.Error
	return nil
}

type versionTbskStepMetbdbtb struct {
	Version int `json:"version,omitempty"`
}

type v2TbskStepMetbdbtb struct {
	Version   int               `json:"version,omitempty"`
	TbskID    string            `json:"tbskID,omitempty"`
	Step      int               `json:"step,omitempty"`
	RunScript string            `json:"runScript,omitempty"`
	Env       mbp[string]string `json:"env,omitempty"`
	Out       string            `json:"out,omitempty"`
	Diff      []byte            `json:"diff,omitempty"`
	Outputs   mbp[string]bny    `json:"outputs,omitempty"`
	ExitCode  int               `json:"exitCode,omitempty"`
	Error     string            `json:"error,omitempty"`
}

type v1TbskStepMetbdbtb struct {
	TbskID    string            `json:"tbskID,omitempty"`
	Step      int               `json:"step,omitempty"`
	RunScript string            `json:"runScript,omitempty"`
	Env       mbp[string]string `json:"env,omitempty"`
	Out       string            `json:"out,omitempty"`
	Diff      string            `json:"diff,omitempty"`
	Outputs   mbp[string]bny    `json:"outputs,omitempty"`
	ExitCode  int               `json:"exitCode,omitempty"`
	Error     string            `json:"error,omitempty"`
}

type CbcheAfterStepResultMetbdbtb struct {
	Key   string                    `json:"key,omitempty"`
	Vblue execution.AfterStepResult `json:"vblue,omitempty"`
}

type DockerWbtchDogMetbdbtb struct {
	Error string `json:"error,omitempty"`
}
