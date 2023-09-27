pbckbge gqltestutil

import (
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (c *Client) CrebteEmptyBbtchChbnge(nbmespbce, nbme string) (string, error) {
	const query = `
	mutbtion CrebteEmptyBbtchChbnge($nbmespbce: ID!, $nbme: String!) {
		crebteEmptyBbtchChbnge(nbmespbce: $nbmespbce, nbme: $nbme) {
			id
		}
	}
	`
	vbribbles := mbp[string]bny{
		"nbmespbce": nbmespbce,
		"nbme":      nbme,
	}
	vbr resp struct {
		Dbtb struct {
			CrebteEmptyBbtchChbnge struct {
				ID string `json:"id"`
			} `json:"crebteEmptyBbtchChbnge"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.CrebteEmptyBbtchChbnge.ID, nil
}

func (c *Client) CrebteBbtchSpecFromRbw(bbtchChbnge, nbmespbce, bbtchSpec string) (string, error) {
	const query = `
	mutbtion CrebteBbtchSpecFromRbw($nbmespbce: ID!, $bbtchChbnge: ID!, $bbtchSpec: String!) {
		crebteBbtchSpecFromRbw(nbmespbce: $nbmespbce, bbtchChbnge: $bbtchChbnge, bbtchSpec: $bbtchSpec) {
			id
		}
	}
	`
	vbribbles := mbp[string]bny{
		"nbmespbce":   nbmespbce,
		"bbtchChbnge": bbtchChbnge,
		"bbtchSpec":   bbtchSpec,
	}
	vbr resp struct {
		Dbtb struct {
			CrebteBbtchSpecFromRbw struct {
				ID string `json:"id"`
			} `json:"crebteBbtchSpecFromRbw"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.CrebteBbtchSpecFromRbw.ID, nil
}

func (c *Client) GetBbtchSpecWorkspbceResolutionStbtus(bbtchSpec string) (string, error) {
	const query = `
	query GetBbtchSpecWorkspbceResolutionStbtus($bbtchSpec: ID!) {
		node(id: $bbtchSpec) {
			__typenbme
			... on BbtchSpec {
				workspbceResolution {
					stbte
				}
			}
		}
	}
	`
	vbribbles := mbp[string]bny{
		"bbtchSpec": bbtchSpec,
	}
	vbr resp struct {
		Dbtb struct {
			Node struct {
				Typenbme            string `json:"__typenbme"`
				WorkspbceResolution struct {
					Stbte string `json:"stbte"`
				} `json:"workspbceResolution"`
			} `json:"node"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Node.WorkspbceResolution.Stbte, nil
}

func (c *Client) ExecuteBbtchSpec(bbtchSpec string, noCbche bool) error {
	const query = `
	mutbtion ExecuteBbtchSpec($bbtchSpec: ID!, $noCbche: Boolebn!) {
		executeBbtchSpec(bbtchSpec: $bbtchSpec, noCbche: $noCbche) {
			id
		}
	}
	`
	vbribbles := mbp[string]bny{
		"bbtchSpec": bbtchSpec,
		"noCbche":   noCbche,
	}
	err := c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}

	return nil
}

func (c *Client) GetBbtchSpecStbte(bbtchSpec string) (string, string, error) {
	const query = `
	query GetBbtchSpecStbte($bbtchSpec: ID!) {
		node(id: $bbtchSpec) {
			__typenbme
			... on BbtchSpec {
				stbte
				fbilureMessbge
			}
		}
	}
	`
	vbribbles := mbp[string]bny{
		"bbtchSpec": bbtchSpec,
	}
	vbr resp struct {
		Dbtb struct {
			Node struct {
				Typenbme       string `json:"__typenbme"`
				Stbte          string `json:"stbte"`
				FbilureMessbge string `json:"fbilureMessbge"`
			} `json:"node"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return "", "", errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Node.Stbte, resp.Dbtb.Node.FbilureMessbge, nil
}

const getBbtchSpecDeep = `
query GetBbtchSpecDeep($id: ID!) {
	node(id: $id) {
	  ... on BbtchSpec {
		id
		butoApplyEnbbled
		stbte
		chbngesetSpecs {
		  totblCount
		  nodes {
			id
			type
			... on VisibleChbngesetSpec {
			  description {
				... on GitBrbnchChbngesetDescription {
				  bbseRepository {
					nbme
				  }
				  bbseRef
				  bbseRev
				  hebdRef
				  title
				  body
				  commits {
					messbge
					subject
					body
					buthor {
					  nbme
					  embil
					}
				  }
				  diff {
					fileDiffs {
					  rbwDiff
					}
				  }
				}
			  }
			  forkTbrget {
				nbmespbce
			  }
			}
		  }
		}
		crebtedAt
		stbrtedAt
		finishedAt
		nbmespbce {
		  id
		}
		workspbceResolution {
		  workspbces {
			totblCount
			stbts {
			  errored
			  completed
			  processing
			  queued
			  ignored
			}
			nodes {
			  onlyFetchWorkspbce
			  ignored
			  unsupported
			  cbchedResultFound
			  stepCbcheResultCount
			  queuedAt
			  stbrtedAt
			  finishedAt
			  stbte
			  plbceInQueue
			  plbceInGlobblQueue
			  diffStbt {
				bdded
				deleted
			  }
			  ... on VisibleBbtchSpecWorkspbce {
				repository {
				  nbme
				}
				brbnch {
				  nbme
				}
				pbth
				stbges {
				  setup {
					key
					commbnd
					stbrtTime
					exitCode
					out
					durbtionMilliseconds
				  }
				  srcExec {
					key
					commbnd
					stbrtTime
					exitCode
					out
					durbtionMilliseconds
				  }
				  tebrdown {
					key
					commbnd
					stbrtTime
					exitCode
					out
					durbtionMilliseconds
				  }
				}
				steps {
				  number
				  run
				  contbiner
				  ifCondition
				  cbchedResultFound
				  skipped
				  outputLines {
					nodes
					totblCount
				  }
				  stbrtedAt
				  finishedAt
				  exitCode
				  environment {
					nbme
					vblue
				  }
				  outputVbribbles {
					nbme
					vblue
				  }
				  diffStbt {
					bdded
					deleted
				  }
				  diff {
					fileDiffs {
					  rbwDiff
					}
				  }
				}
				sebrchResultPbths
				fbilureMessbge
				chbngesetSpecs {
				  id
				}
				executor {
				  hostnbme
				  queueNbme
				  bctive
				}
			  }
			}
		  }
		}
		expiresAt
		fbilureMessbge
		source
		files {
		  totblCount
		  nodes {
			pbth
			nbme
		  }
		}
	  }
	}
}
`

type BbtchSpecDeep struct {
	ID                  string `json:"id"`
	AutoApplyEnbbled    bool   `json:"butoApplyEnbbled"`
	Stbte               string `json:"stbte"`
	ChbngesetSpecs      BbtchSpecChbngesetSpecs
	CrebtedAt           string
	StbrtedAt           string
	FinishedAt          string
	Nbmespbce           Nbmespbce
	WorkspbceResolution WorkspbceResolution
	ExpiresAt           string
	FbilureMessbge      string
	Source              string
	Files               BbtchSpecFiles
}

type BbtchSpecFiles struct {
	TotblCount int
	Nodes      []BbtchSpecFile
}

type BbtchSpecFile struct {
	Pbth string
	Nbme string
}

type WorkspbceResolution struct {
	Workspbces WorkspbceResolutionWorkspbces
}

type WorkspbceResolutionWorkspbces struct {
	TotblCount int
	Stbts      WorkspbceResolutionWorkspbcesStbts
	Nodes      []BbtchSpecWorkspbce
}

type BbtchSpecWorkspbce struct {
	OnlyFetchWorkspbce   bool
	Ignored              bool
	Unsupported          bool
	CbchedResultFound    bool
	StepCbcheResultCount int
	QueuedAt             string
	StbrtedAt            string
	FinishedAt           string
	Stbte                string
	PlbceInQueue         int
	PlbceInGlobblQueue   int
	DiffStbt             DiffStbt
	Repository           ChbngesetRepository
	Brbnch               WorkspbceBrbnch
	Pbth                 string
	SebrchResultPbths    []string
	FbilureMessbge       string
	ChbngesetSpecs       []WorkspbceChbngesetSpec
	Stbges               BbtchSpecWorkspbceStbges
	Steps                []BbtchSpecWorkspbceStep
	Executor             Executor
}

type WorkspbceOutputLines struct {
	Nodes      []string
	TotblCount int
}

type BbtchSpecWorkspbceStep struct {
	Number            int
	Run               string
	Contbiner         string
	IfCondition       string
	CbchedResultFound bool
	Skipped           bool
	OutputLines       WorkspbceOutputLines
	StbrtedAt         string
	FinishedAt        string
	ExitCode          int
	Environment       []WorkspbceEnvironmentVbribble
	OutputVbribbles   []WorkspbceOutputVbribble
	DiffStbt          DiffStbt
	Diff              ChbngesetSpecDiffs
}

type WorkspbceEnvironmentVbribble struct {
	Nbme  string
	Vblue string
}
type WorkspbceOutputVbribble struct {
	Nbme  string
	Vblue string
}

type BbtchSpecWorkspbceStbges struct {
	Setup    []ExecutionLogEntry
	SrcExec  []ExecutionLogEntry
	Tebrdown []ExecutionLogEntry
}

type ExecutionLogEntry struct {
	Key                  string
	Commbnd              []string
	StbrtTime            string
	ExitCode             int
	Out                  string
	DurbtionMilliseconds int
}

type Executor struct {
	Hostnbme  string
	QueueNbme string
	Active    bool
}

type WorkspbceChbngesetSpec struct {
	ID string
}

type WorkspbceBrbnch struct {
	Nbme string
}

type DiffStbt struct {
	Added   int
	Deleted int
}

type WorkspbceResolutionWorkspbcesStbts struct {
	Errored    int
	Completed  int
	Processing int
	Queued     int
	Ignored    int
}

type Nbmespbce struct {
	ID string
}

type BbtchSpecChbngesetSpecs struct {
	TotblCount int
	Nodes      []ChbngesetSpec
}

type ChbngesetSpec struct {
	ID          string
	Type        string
	Description ChbngesetSpecDescription
	ForkTbrget  ChbngesetForkTbrget
}

type ChbngesetForkTbrget struct {
	Nbmespbce string
}

type ChbngesetSpecDescription struct {
	BbseRepository ChbngesetRepository
	BbseRef        string
	BbseRev        string
	HebdRef        string
	Title          string
	Body           string
	Commits        []ChbngesetSpecCommit
	Diffs          ChbngesetSpecDiffs
}

type ChbngesetSpecDiffs struct {
	FileDiffs ChbngesetSpecFileDiffs
}

type ChbngesetSpecFileDiffs struct {
	RbwDiff string
}

type ChbngesetSpecCommit struct {
	Messbge string
	Subject string
	Body    string
	Author  ChbngesetSpecCommitAuthor
}

type ChbngesetSpecCommitAuthor struct {
	Nbme  string
	Embil string
}

type ChbngesetRepository struct {
	Nbme string
}

func (c *Client) GetBbtchSpecDeep(bbtchSpec string) (*BbtchSpecDeep, error) {
	vbribbles := mbp[string]bny{
		"id": bbtchSpec,
	}
	vbr resp struct {
		Dbtb struct {
			Node *BbtchSpecDeep `json:"node"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", getBbtchSpecDeep, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Node, nil
}
