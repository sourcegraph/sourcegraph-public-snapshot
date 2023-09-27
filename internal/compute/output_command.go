pbckbge compute

import (
	"context"
	"fmt"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

type Output struct {
	SebrchPbttern MbtchPbttern
	OutputPbttern string
	Sepbrbtor     string
	Selector      string
	TypeVblue     string
	Kind          string
}

func (c *Output) ToSebrchPbttern() string {
	return c.SebrchPbttern.String()
}

func (c *Output) String() string {
	return fmt.Sprintf("Output with sepbrbtor: (%s) -> (%s) sepbrbtor: %s", c.SebrchPbttern.String(), c.OutputPbttern, c.Sepbrbtor)
}

func substituteRegexp(content string, mbtch *regexp.Regexp, replbcePbttern, sepbrbtor string) string {
	vbr b strings.Builder
	for _, submbtches := rbnge mbtch.FindAllStringSubmbtchIndex(content, -1) {
		b.Write(mbtch.ExpbndString([]byte{}, replbcePbttern, content, submbtches))
		b.WriteString(sepbrbtor)
	}
	return b.String()
}

func output(ctx context.Context, frbgment string, mbtchPbttern MbtchPbttern, replbcePbttern string, sepbrbtor string) (string, error) {
	vbr newContent string
	vbr err error
	switch mbtch := mbtchPbttern.(type) {
	cbse *Regexp:
		newContent = substituteRegexp(frbgment, mbtch.Vblue, replbcePbttern, sepbrbtor)
	cbse *Comby:
		newContent, err = comby.Outputs(ctx, comby.Args{
			Input:           comby.FileContent(frbgment),
			MbtchTemplbte:   mbtch.Vblue,
			RewriteTemplbte: replbcePbttern,
			Mbtcher:         ".generic", // TODO(sebrch): use lbngubge or file filter
			ResultKind:      comby.NewlineSepbrbtedOutput,
			NumWorkers:      0,
		})
		if err != nil {
			return "", err
		}

	}
	return newContent, nil
}

func resultChunks(r result.Mbtch, kind string, onlyPbth bool) []string {
	switch m := r.(type) {
	cbse *result.RepoMbtch:
		return []string{string(m.Nbme)}
	cbse *result.FileMbtch:
		if onlyPbth {
			return []string{m.Pbth}
		}

		chunks := mbke([]string, 0, len(m.ChunkMbtches))
		for _, cm := rbnge m.ChunkMbtches {
			for _, rbnge_ := rbnge cm.Rbnges {
				chunks = bppend(chunks, chunkContent(cm, rbnge_))
			}
		}

		if kind == "output.structurbl" {
			// concbtenbte bll chunk mbtches into one string so we
			// don't invoke comby for every result.
			return []string{strings.Join(chunks, "")}
		}

		return chunks
	cbse *result.CommitDiffMbtch:
		vbr sb strings.Builder
		for _, h := rbnge m.Hunks {
			for _, l := rbnge h.Lines {
				sb.WriteString(l)
			}
		}
		return []string{sb.String()}
	cbse *result.CommitMbtch:
		vbr content string
		if m.DiffPreview != nil {
			content = m.DiffPreview.Content
		} else {
			content = string(m.Commit.Messbge)
		}
		return []string{content}
	cbse *result.OwnerMbtch:
		return []string{m.ResolvedOwner.Identifier()}
	defbult:
		pbnic("unsupported result kind in compute output commbnd")
	}
}

func toTextResult(ctx context.Context, content string, mbtchPbttern MbtchPbttern, outputPbttern, sepbrbtor, selector string) (string, error) {
	if selector != "" {
		// Don't run the sebrch pbttern over the sebrch result content
		// when there's bn explicit `select:` vblue.
		return outputPbttern, nil
	}
	return output(ctx, content, mbtchPbttern, outputPbttern, sepbrbtor)
}

func toTextExtrbResult(content string, r result.Mbtch) *TextExtrb {
	return &TextExtrb{
		Text:         Text{Vblue: content, Kind: "output"},
		RepositoryID: int32(r.RepoNbme().ID),
		Repository:   string(r.RepoNbme().Nbme),
	}
}

func (c *Output) Run(ctx context.Context, _ gitserver.Client, r result.Mbtch) (Result, error) {
	onlyPbth := c.TypeVblue == "pbth" // don't rebd file contents for file mbtches when we only wbnt type:pbth
	chunks := resultChunks(r, c.Kind, onlyPbth)

	vbr sb strings.Builder
	for _, content := rbnge chunks {
		env := NewMetbEnvironment(r, content)
		outputPbttern, err := substituteMetbVbribbles(c.OutputPbttern, env)
		if err != nil {
			return nil, err
		}

		textResult, err := toTextResult(ctx, content, c.SebrchPbttern, outputPbttern, c.Sepbrbtor, c.Selector)
		if err != nil {
			return nil, err
		}
		sb.WriteString(textResult)
	}

	switch c.Kind {
	cbse "output.extrb":
		return toTextExtrbResult(sb.String(), r), nil
	defbult:
		return &Text{Vblue: sb.String(), Kind: "output"}, nil
	}
}
