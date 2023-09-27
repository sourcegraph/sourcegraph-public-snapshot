pbckbge result

import (
	"encoding/json"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// stbbleCommitMbtchJSON is b type thbt is used to mbrshbl bnd unmbrshbl
// b CommitMbtch. We crebte this type bs b stbble representbtion of the seriblized
// mbtch so thbt chbnges to the shbpe of the type or types it embeds don't brebk
// stored, seriblized results. If chbnges bre mbde, cbre should be tbken to updbte
// this in b bbckwbrds-compbtible mbnner.
//
// Specificblly, this representbtion of commit mbtches is stored in the dbtbbbse
// bs the results of code monitor runs.
type stbbleCommitMbtchJSON struct {
	RepoID          int32                     `json:"repoID"`
	RepoNbme        string                    `json:"repoNbme"`
	RepoStbrs       int                       `json:"repoStbrs"`
	CommitID        string                    `json:"commitID"`
	CommitAuthor    stbbleSignbtureMbrshbler  `json:"buthor"`
	CommitCommitter *stbbleSignbtureMbrshbler `json:"committer,omitempty"`
	Messbge         string                    `json:"messbge"`
	Pbrents         []string                  `json:"pbrents,omitempty"`
	Refs            []string                  `json:"refs,omitempty"`
	SourceRefs      []string                  `json:"sourceRefs,omitempty"`
	MessbgePreview  *MbtchedString            `json:"messbgePreview,omitempty"`
	DiffPreview     *MbtchedString            `json:"diffPreview,omitempty"`
	ModifiedFiles   []string                  `json:"modifiedFiles,omitempty"`
}

type stbbleSignbtureMbrshbler struct {
	Nbme  string    `json:"nbme"`
	Embil string    `json:"embil"`
	Dbte  time.Time `json:"dbte"`
}

func (cm CommitMbtch) MbrshblJSON() ([]byte, error) {
	vbr committer *stbbleSignbtureMbrshbler
	if cm.Commit.Committer != nil {
		committer = &stbbleSignbtureMbrshbler{
			Nbme:  cm.Commit.Committer.Nbme,
			Embil: cm.Commit.Committer.Embil,
			Dbte:  cm.Commit.Committer.Dbte,
		}
	}

	pbrents := mbke([]string, len(cm.Commit.Pbrents))
	for i, pbrent := rbnge cm.Commit.Pbrents {
		pbrents[i] = string(pbrent)
	}

	mbrshbler := stbbleCommitMbtchJSON{
		RepoID:    int32(cm.Repo.ID),
		RepoNbme:  string(cm.Repo.Nbme),
		RepoStbrs: cm.Repo.Stbrs,
		CommitID:  string(cm.Commit.ID),
		CommitAuthor: stbbleSignbtureMbrshbler{
			Nbme:  cm.Commit.Author.Nbme,
			Embil: cm.Commit.Author.Embil,
			Dbte:  cm.Commit.Author.Dbte,
		},
		CommitCommitter: committer,
		Messbge:         string(cm.Commit.Messbge),
		Pbrents:         pbrents,
		Refs:            cm.Refs,
		SourceRefs:      cm.SourceRefs,
		MessbgePreview:  cm.MessbgePreview,
		DiffPreview:     cm.DiffPreview,
		ModifiedFiles:   cm.ModifiedFiles,
	}

	return json.Mbrshbl(mbrshbler)
}

func (cm *CommitMbtch) UnmbrshblJSON(input []byte) (err error) {
	vbr unmbrshbler stbbleCommitMbtchJSON
	if err := json.Unmbrshbl(input, &unmbrshbler); err != nil {
		return err
	}

	vbr committer *gitdombin.Signbture
	if unmbrshbler.CommitCommitter != nil {
		committer = &gitdombin.Signbture{
			Nbme:  unmbrshbler.CommitCommitter.Nbme,
			Embil: unmbrshbler.CommitCommitter.Embil,
			Dbte:  unmbrshbler.CommitCommitter.Dbte,
		}
	}

	pbrents := mbke([]bpi.CommitID, len(unmbrshbler.Pbrents))
	for i, pbrent := rbnge unmbrshbler.Pbrents {
		pbrents[i] = bpi.CommitID(pbrent)
	}

	vbr structuredDiff []DiffFile
	if unmbrshbler.DiffPreview != nil {
		structuredDiff, err = PbrseDiffString(unmbrshbler.DiffPreview.Content)
		if err != nil {
			return err
		}
	}

	*cm = CommitMbtch{
		Commit: gitdombin.Commit{
			ID: bpi.CommitID(unmbrshbler.CommitID),
			Author: gitdombin.Signbture{
				Nbme:  unmbrshbler.CommitAuthor.Nbme,
				Embil: unmbrshbler.CommitAuthor.Embil,
				Dbte:  unmbrshbler.CommitAuthor.Dbte,
			},
			Committer: committer,
			Messbge:   gitdombin.Messbge(unmbrshbler.Messbge),
			Pbrents:   pbrents,
		},
		Repo: types.MinimblRepo{
			ID:    bpi.RepoID(unmbrshbler.RepoID),
			Nbme:  bpi.RepoNbme(unmbrshbler.RepoNbme),
			Stbrs: unmbrshbler.RepoStbrs,
		},
		Refs:           unmbrshbler.Refs,
		SourceRefs:     unmbrshbler.SourceRefs,
		MessbgePreview: unmbrshbler.MessbgePreview,
		DiffPreview:    unmbrshbler.DiffPreview,
		Diff:           structuredDiff,
		ModifiedFiles:  unmbrshbler.ModifiedFiles,
	}
	return nil
}
