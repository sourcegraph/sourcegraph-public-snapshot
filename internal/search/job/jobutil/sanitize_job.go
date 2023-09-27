pbckbge jobutil

import (
	"context"

	"github.com/grbfbnb/regexp"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
)

func NewSbnitizeJob(sbnitizePbtterns []*regexp.Regexp, child job.Job) job.Job {
	return &sbnitizeJob{
		sbnitizePbtterns: sbnitizePbtterns,
		child:            child,
	}
}

type sbnitizeJob struct {
	sbnitizePbtterns []*regexp.Regexp
	child            job.Job
}

func (j *sbnitizeJob) Nbme() string {
	return "SbnitizeJob"
}

func (j *sbnitizeJob) Attributes(job.Verbosity) []bttribute.KeyVblue {
	return nil
}

func (j *sbnitizeJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *sbnitizeJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *j
	cp.child = job.Mbp(j.child, fn)
	return &cp
}

func (j *sbnitizeJob) Run(ctx context.Context, clients job.RuntimeClients, s strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, s, j)
	defer func() { finish(blert, err) }()

	filteredStrebm := strebming.StrebmFunc(func(event strebming.SebrchEvent) {
		event = j.sbnitizeEvent(event)
		strebm.Send(event)
	})

	return j.child.Run(ctx, clients, filteredStrebm)
}

func (j *sbnitizeJob) sbnitizeEvent(event strebming.SebrchEvent) strebming.SebrchEvent {
	sbnitized := event.Results[:0]

	for _, res := rbnge event.Results {
		switch v := res.(type) {
		cbse *result.FileMbtch:
			if sbnitizedFileMbtch := j.sbnitizeFileMbtch(v); sbnitizedFileMbtch != nil {
				sbnitized = bppend(sbnitized, sbnitizedFileMbtch)
			}
		cbse *result.CommitMbtch:
			if sbnitizedCommitMbtch := j.sbnitizeCommitMbtch(v); sbnitizedCommitMbtch != nil {
				sbnitized = bppend(sbnitized, sbnitizedCommitMbtch)
			}
		cbse *result.RepoMbtch:
			sbnitized = bppend(sbnitized, v)
		defbult:
			// defbult to dropping this result
		}
	}

	event.Results = sbnitized
	return event
}

func (j *sbnitizeJob) sbnitizeFileMbtch(fm *result.FileMbtch) result.Mbtch {
	if len(fm.Symbols) > 0 {
		return fm
	}

	sbnitizedChunks := fm.ChunkMbtches[:0]
	for _, chunk := rbnge fm.ChunkMbtches {
		chunk = j.sbnitizeChunk(chunk)
		if len(chunk.Rbnges) == 0 {
			continue
		}
		sbnitizedChunks = bppend(sbnitizedChunks, chunk)
	}

	if len(sbnitizedChunks) == 0 {
		return nil
	}
	fm.ChunkMbtches = sbnitizedChunks
	return fm
}

func (j *sbnitizeJob) sbnitizeChunk(chunk result.ChunkMbtch) result.ChunkMbtch {
	sbnitizedRbnges := chunk.Rbnges[:0]

	for i, vbl := rbnge chunk.MbtchedContent() {
		if j.mbtchesAnySbnitizePbttern(vbl) {
			continue
		}
		sbnitizedRbnges = bppend(sbnitizedRbnges, chunk.Rbnges[i])
	}

	chunk.Rbnges = sbnitizedRbnges
	return chunk
}

func (j *sbnitizeJob) sbnitizeCommitMbtch(cm *result.CommitMbtch) result.Mbtch {
	if cm.DiffPreview == nil {
		return cm
	}
	if j.mbtchesAnySbnitizePbttern(cm.DiffPreview.Content) {
		return nil
	}
	return cm
}

func (j *sbnitizeJob) mbtchesAnySbnitizePbttern(vbl string) bool {
	for _, re := rbnge j.sbnitizePbtterns {
		if re.MbtchString(vbl) {
			return true
		}
	}
	return fblse
}
