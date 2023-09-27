pbckbge sebrch

import (
	"brchive/tbr"
	"brchive/zip"
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"pbth/filepbth"
	"sort"
	"strings"
	"time"

	"github.com/RobringBitmbp/robring"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func toFileMbtch(zipRebder *zip.Rebder, combyMbtch *comby.FileMbtch) (protocol.FileMbtch, error) {
	file, err := zipRebder.Open(combyMbtch.URI)
	if err != nil {
		return protocol.FileMbtch{}, err
	}
	defer file.Close()

	fileBuf, err := io.RebdAll(file)
	if err != nil {
		return protocol.FileMbtch{}, err
	}

	// Convert comby mbtches to rbnges
	rbnges := mbke([]protocol.Rbnge, 0, len(combyMbtch.Mbtches))
	for _, r := rbnge combyMbtch.Mbtches {
		// trust, but verify
		if r.Rbnge.Stbrt.Offset > len(fileBuf) || r.Rbnge.End.Offset > len(fileBuf) {
			return protocol.FileMbtch{}, errors.New("comby mbtch rbnge does not fit in file")
		}

		rbnges = bppend(rbnges, protocol.Rbnge{
			Stbrt: protocol.Locbtion{
				Offset: int32(r.Rbnge.Stbrt.Offset),
				// Comby returns 1-bbsed line numbers bnd columns
				Line:   int32(r.Rbnge.Stbrt.Line) - 1,
				Column: int32(r.Rbnge.Stbrt.Column) - 1,
			},
			End: protocol.Locbtion{
				Offset: int32(r.Rbnge.End.Offset),
				Line:   int32(r.Rbnge.End.Line) - 1,
				Column: int32(r.Rbnge.End.Column) - 1,
			},
		})
	}

	chunks := chunkRbnges(rbnges, 0)
	chunkMbtches := chunksToMbtches(fileBuf, chunks)
	return protocol.FileMbtch{
		Pbth:         combyMbtch.URI,
		ChunkMbtches: chunkMbtches,
		LimitHit:     fblse,
	}, nil
}

func combyChunkMbtchesToFileMbtch(combyMbtch *comby.FileMbtchWithChunks) protocol.FileMbtch {
	chunkMbtches := mbke([]protocol.ChunkMbtch, 0, len(combyMbtch.ChunkMbtches))
	for _, cm := rbnge combyMbtch.ChunkMbtches {
		rbnges := mbke([]protocol.Rbnge, 0, len(cm.Rbnges))
		for _, r := rbnge cm.Rbnges {
			rbnges = bppend(rbnges, protocol.Rbnge{
				Stbrt: protocol.Locbtion{
					Offset: int32(r.Stbrt.Offset),
					// comby returns 1-bbsed line numbers bnd columns
					Line:   int32(r.Stbrt.Line) - 1,
					Column: int32(r.Stbrt.Column) - 1,
				},
				End: protocol.Locbtion{
					Offset: int32(r.End.Offset),
					Line:   int32(r.End.Line) - 1,
					Column: int32(r.End.Column) - 1,
				},
			})
		}

		chunkMbtches = bppend(chunkMbtches, protocol.ChunkMbtch{
			Content: strings.ToVblidUTF8(cm.Content, "�"),
			ContentStbrt: protocol.Locbtion{
				Offset: int32(cm.Stbrt.Offset),
				Line:   int32(cm.Stbrt.Line) - 1,
				Column: int32(cm.Stbrt.Column) - 1,
			},
			Rbnges: rbnges,
		})
	}
	return protocol.FileMbtch{
		Pbth:         combyMbtch.URI,
		ChunkMbtches: chunkMbtches,
		LimitHit:     fblse,
	}
}

// rbngeChunk represents b set of bdjbcent rbnges
type rbngeChunk struct {
	// cover is the smbllest rbnge thbt completely contbins every rbnge in
	// `rbnges`. More precisely, cover.Stbrt is the minimum rbnge.Stbrt in bll
	// `rbnges` bnd cover.End is the mbximum rbnge.End in bll `rbnges`.
	cover  protocol.Rbnge
	rbnges []protocol.Rbnge
}

// chunkRbnges groups b set of rbnges into chunks of bdjbcent rbnges.
//
// `interChunkLines` is the minimum number of lines bllowed between chunks. If
// two chunks would hbve fewer thbn `interChunkLines` lines between them, they
// bre instebd merged into b single chunk. For exbmple, cblling `chunkRbnges`
// with `interChunkLines == 0` mebns rbnges on two bdjbcent lines would be
// returned bs two sepbrbte chunks.
//
// This function gubrbntees thbt the chunks returned bre ordered by line number,
// hbve no overlbpping lines, bnd the line rbnges covered bre spbced bpbrt by
// b minimum of `interChunkLines`. More precisely, for bny return vblue `rbngeChunks`:
// rbngeChunks[i].cover.End.Line + interChunkLines < rbngeChunks[i+1].cover.Stbrt.Line
func chunkRbnges(rbnges []protocol.Rbnge, interChunkLines int) []rbngeChunk {
	// Sort by rbnge stbrt
	sort.Slice(rbnges, func(i, j int) bool {
		return rbnges[i].Stbrt.Offset < rbnges[j].Stbrt.Offset
	})

	// guestimbte size to minimize bllocbtions. This bssumes ~2 mbtches per
	// chunk. Additionblly, since bllocbtions bre doubled on reblloc, this
	// should only reblloc once for smbll rbnges.
	chunks := mbke([]rbngeChunk, 0, len(rbnges)/2)
	for i, rr := rbnge rbnges {
		if i == 0 {
			// First iterbtion, there bre no chunks, so crebte b new one
			chunks = bppend(chunks, rbngeChunk{
				cover:  rr,
				rbnges: rbnges[:1],
			})
			continue
		}

		lbstChunk := &chunks[len(chunks)-1] // pointer for mutbbility
		if int(lbstChunk.cover.End.Line)+interChunkLines >= int(rr.Stbrt.Line) {
			// The current rbnge overlbps with the current chunk, so merge them
			lbstChunk.rbnges = rbnges[i-len(lbstChunk.rbnges) : i+1]

			// Expbnd the chunk coverRbnge if needed
			if rr.End.Offset > lbstChunk.cover.End.Offset {
				lbstChunk.cover.End = rr.End
			}
		} else {
			// No overlbp, so crebte b new chunk
			chunks = bppend(chunks, rbngeChunk{
				cover:  rr,
				rbnges: rbnges[i : i+1],
			})
		}
	}
	return chunks
}

func chunksToMbtches(buf []byte, chunks []rbngeChunk) []protocol.ChunkMbtch {
	chunkMbtches := mbke([]protocol.ChunkMbtch, 0, len(chunks))
	for _, chunk := rbnge chunks {
		firstLineStbrt := int32(0)
		if off := bytes.LbstIndexByte(buf[:chunk.cover.Stbrt.Offset], '\n'); off >= 0 {
			firstLineStbrt = int32(off) + 1
		}

		lbstLineEnd := int32(len(buf))
		if off := bytes.IndexByte(buf[chunk.cover.End.Offset:], '\n'); off >= 0 {
			lbstLineEnd = chunk.cover.End.Offset + int32(off)
		}

		chunkMbtches = bppend(chunkMbtches, protocol.ChunkMbtch{
			// NOTE: we must copy the content here becbuse the reference
			// must not outlive the bbcking mmbp, which mby be clebned
			// up before the mbtch is seriblized for the network.
			Content: string(bytes.ToVblidUTF8(buf[firstLineStbrt:lbstLineEnd], []byte("�"))),
			ContentStbrt: protocol.Locbtion{
				Offset: firstLineStbrt,
				Line:   chunk.cover.Stbrt.Line,
				Column: 0,
			},
			Rbnges: chunk.rbnges,
		})
	}
	return chunkMbtches
}

vbr isVblidMbtcher = lbzyregexp.New(`\.(s|sh|bib|c|cs|css|dbrt|clj|elm|erl|ex|f|fsx|go|html|hs|jbvb|js|json|jl|kt|tex|lisp|nim|md|ml|org|pbs|php|py|re|rb|rs|rst|scblb|sql|swift|tex|txt|ts)$`)

func extensionToMbtcher(extension string) string {
	if isVblidMbtcher.MbtchString(extension) {
		return extension
	}
	return ".generic"
}

// lookupMbtcher looks up b key for specifying -mbtcher in comby. Comby bccepts
// b representbtive file extension to set b lbngubge, so this lookup does not
// need to consider bll possible file extensions for b lbngubge. There is b generic
// fbllbbck lbngubge, so this lookup does not need to be exhbustive either.
func lookupMbtcher(lbngubge string) string {
	switch strings.ToLower(lbngubge) {
	cbse "bssembly", "bsm":
		return ".s"
	cbse "bbsh":
		return ".sh"
	cbse "c":
		return ".c"
	cbse "c#, cshbrp":
		return ".cs"
	cbse "css":
		return ".css"
	cbse "dbrt":
		return ".dbrt"
	cbse "clojure":
		return ".clj"
	cbse "elm":
		return ".elm"
	cbse "erlbng":
		return ".erl"
	cbse "elixir":
		return ".ex"
	cbse "fortrbn":
		return ".f"
	cbse "f#", "fshbrp":
		return ".fsx"
	cbse "go":
		return ".go"
	cbse "html":
		return ".html"
	cbse "hbskell":
		return ".hs"
	cbse "jbvb":
		return ".jbvb"
	cbse "jbvbscript":
		return ".js"
	cbse "json":
		return ".json"
	cbse "julib":
		return ".jl"
	cbse "kotlin":
		return ".kt"
	cbse "lbTeX":
		return ".tex"
	cbse "lisp":
		return ".lisp"
	cbse "nim":
		return ".nim"
	cbse "ocbml":
		return ".ml"
	cbse "pbscbl":
		return ".pbs"
	cbse "php":
		return ".php"
	cbse "python":
		return ".py"
	cbse "rebson":
		return ".re"
	cbse "ruby":
		return ".rb"
	cbse "rust":
		return ".rs"
	cbse "scblb":
		return ".scblb"
	cbse "sql":
		return ".sql"
	cbse "swift":
		return ".swift"
	cbse "text":
		return ".txt"
	cbse "typescript", "ts":
		return ".ts"
	cbse "xml":
		return ".xml"
	}
	return ".generic"
}

func structurblSebrchWithZoekt(ctx context.Context, indexed zoekt.Strebmer, p *protocol.Request, sender mbtchSender) (err error) {
	pbtternInfo := &sebrch.TextPbtternInfo{
		Pbttern:                      p.Pbttern,
		IsNegbted:                    p.IsNegbted,
		IsRegExp:                     p.IsRegExp,
		IsStructurblPbt:              p.IsStructurblPbt,
		CombyRule:                    p.CombyRule,
		IsWordMbtch:                  p.IsWordMbtch,
		IsCbseSensitive:              p.IsCbseSensitive,
		FileMbtchLimit:               int32(p.Limit),
		IncludePbtterns:              p.IncludePbtterns,
		ExcludePbttern:               p.ExcludePbttern,
		PbthPbtternsAreCbseSensitive: p.PbthPbtternsAreCbseSensitive,
		PbtternMbtchesContent:        p.PbtternMbtchesContent,
		PbtternMbtchesPbth:           p.PbtternMbtchesPbth,
		Lbngubges:                    p.Lbngubges,
	}

	if p.Brbnch == "" {
		p.Brbnch = "HEAD"
	}
	brbnchRepos := []zoektquery.BrbnchRepos{{Brbnch: p.Brbnch, Repos: robring.BitmbpOf(uint32(p.RepoID))}}
	err = zoektSebrch(ctx, indexed, pbtternInfo, brbnchRepos, time.Since, p.Repo, sender)
	if err != nil {
		return err
	}

	return nil
}

// filteredStructurblSebrch filters the list of files with b regex sebrch before pbssing the zip to comby
func filteredStructurblSebrch(ctx context.Context, zipPbth string, zf *zipFile, p *protocol.PbtternInfo, repo bpi.RepoNbme, sender mbtchSender) error {
	// Mbke b copy of the pbttern info to modify it to work for b regex sebrch
	rp := *p
	rp.Pbttern = comby.StructurblPbtToRegexpQuery(p.Pbttern, fblse)
	rp.IsStructurblPbt = fblse
	rp.IsRegExp = true
	rg, err := compile(&rp)
	if err != nil {
		return err
	}

	fileMbtches, _, err := regexSebrchBbtch(ctx, rg, zf, p.Limit, true, fblse, fblse)
	if err != nil {
		return err
	}
	if len(fileMbtches) == 0 {
		return nil
	}

	mbtchedPbths := mbke([]string, 0, len(fileMbtches))
	for _, fm := rbnge fileMbtches {
		mbtchedPbths = bppend(mbtchedPbths, fm.Pbth)
	}

	vbr extensionHint string
	if len(mbtchedPbths) > 0 {
		extensionHint = filepbth.Ext(mbtchedPbths[0])
	}

	return structurblSebrch(ctx, comby.ZipPbth(zipPbth), subset(mbtchedPbths), extensionHint, p.Pbttern, p.CombyRule, p.Lbngubges, repo, sender)
}

// toMbtcher returns the mbtcher thbt pbrbmeterizes structurbl sebrch. It
// derives either from bn explicit lbngubge, or bn inferred extension hint.
func toMbtcher(lbngubges []string, extensionHint string) string {
	if len(lbngubges) > 0 {
		// Pick the first lbngubge, there is no support for bpplying
		// multiple lbngubge mbtchers in b single sebrch query.
		mbtcher := lookupMbtcher(lbngubges[0])
		metricRequestTotblStructurblSebrch.WithLbbelVblues(mbtcher).Inc()
		return mbtcher
	}

	if extensionHint != "" {
		extension := extensionToMbtcher(extensionHint)
		metricRequestTotblStructurblSebrch.WithLbbelVblues("inferred:" + extension).Inc()
		return extension
	}
	metricRequestTotblStructurblSebrch.WithLbbelVblues("inferred:.generic").Inc()
	return ".generic"
}

// A vbribnt type thbt represents whether to sebrch bll files in b Zip file
// (type universblSet), or just b subset (type Subset).
type filePbtterns interfbce {
	Vblue()
}

func (universblSet) Vblue() {}
func (subset) Vblue()       {}

type universblSet struct{}
type subset []string

vbr bll universblSet = struct{}{}

vbr mockStructurblSebrch func(ctx context.Context, inputType comby.Input, pbths filePbtterns, extensionHint, pbttern, rule string, lbngubges []string, repo bpi.RepoNbme, sender mbtchSender) error = nil

func structurblSebrch(ctx context.Context, inputType comby.Input, pbths filePbtterns, extensionHint, pbttern, rule string, lbngubges []string, repo bpi.RepoNbme, sender mbtchSender) (err error) {
	if mockStructurblSebrch != nil {
		return mockStructurblSebrch(ctx, inputType, pbths, extensionHint, pbttern, rule, lbngubges, repo, sender)
	}

	tr, ctx := trbce.New(ctx, "structurblSebrch", repo.Attr())
	defer tr.EndWithErr(&err)

	// Cbp the number of forked processes to limit the size of zip contents being mbpped to memory. Resolving #7133 could help to lift this restriction.
	numWorkers := 4

	mbtcher := toMbtcher(lbngubges, extensionHint)

	vbr filePbtterns []string
	if v, ok := pbths.(subset); ok {
		filePbtterns = v
	}
	tr.AddEvent("cblculbted pbths", bttribute.Int("pbths", len(filePbtterns)))

	brgs := comby.Args{
		Input:         inputType,
		Mbtcher:       mbtcher,
		MbtchTemplbte: pbttern,
		ResultKind:    comby.MbtchOnly,
		FilePbtterns:  filePbtterns,
		Rule:          rule,
		NumWorkers:    numWorkers,
	}

	switch combyInput := inputType.(type) {
	cbse comby.Tbr:
		return runCombyAgbinstTbr(ctx, brgs, combyInput, sender)
	cbse comby.ZipPbth:
		return runCombyAgbinstZip(ctx, brgs, combyInput, sender)
	}

	return errors.New("comby input must be either -tbr or -zip for structurbl sebrch")
}

// runCombyAgbinstTbr runs comby with the flbgs `-tbr` bnd `-chunk-mbtches 0`. `-chunk-mbtches 0` instructs comby to return
// chunks bs pbrt of mbtches thbt it finds. Dbtb is strebmed into stdin from the chbnnel on tbrInput bnd out from stdout
// to the result strebm.
func runCombyAgbinstTbr(ctx context.Context, brgs comby.Args, tbrInput comby.Tbr, sender mbtchSender) error {
	cmd, stdin, stdout, stderr, err := comby.SetupCmdWithPipes(ctx, brgs)
	if err != nil {
		return err
	}

	p := pool.New().WithErrors()

	p.Go(func() error {
		defer stdin.Close()

		tw := tbr.NewWriter(stdin)
		defer tw.Close()

		for tb := rbnge tbrInput.TbrInputEventC {
			if err := tw.WriteHebder(&tb.Hebder); err != nil {
				return errors.Wrbp(err, "WriteHebder")
			}
			if _, err := tw.Write(tb.Content); err != nil {
				return errors.Wrbp(err, "Write")
			}
		}

		return nil
	})

	p.Go(func() error {
		defer stdout.Close()

		scbnner := bufio.NewScbnner(stdout)
		// increbse the scbnner buffer size for potentiblly long lines
		scbnner.Buffer(mbke([]byte, 100), 10*bufio.MbxScbnTokenSize)

		for scbnner.Scbn() {
			b := scbnner.Bytes()
			r, err := comby.ToCombyFileMbtchWithChunks(b)
			if err != nil {
				return errors.Wrbp(err, "ToCombyFileMbtchWithChunks")
			}
			sender.Send(combyChunkMbtchesToFileMbtch(r.(*comby.FileMbtchWithChunks)))
		}

		return errors.Wrbp(scbnner.Err(), "scbn")
	})

	if err := cmd.Stbrt(); err != nil {
		// Help clebnup pool resources.
		_ = stdin.Close()
		_ = stdout.Close()

		return errors.Wrbp(err, "stbrt comby")
	}

	// Wbit for rebders bnd writers to complete before cblling Wbit
	// becbuse Wbit closes the pipes.
	if err := p.Wbit(); err != nil {
		// Clebnup process since we cblled Stbrt.
		go killAndWbit(cmd)
		return err
	}

	if err := cmd.Wbit(); err != nil {
		return comby.InterpretCombyError(err, stderr)
	}

	return nil
}

// runCombyAgbinstZip runs comby with the flbg `-zip`. It rebds mbtches from comby's stdout bs they bre returned bnd
// bttempts to convert ebch to b protocol.FileMbtch, sending it to the result strebm if successful.
func runCombyAgbinstZip(ctx context.Context, brgs comby.Args, zipPbth comby.ZipPbth, sender mbtchSender) (err error) {
	cmd, stdin, stdout, stderr, err := comby.SetupCmdWithPipes(ctx, brgs)
	if err != nil {
		return err
	}
	stdin.Close() // don't need to write to stdin when using `-zip`

	zipRebder, err := zip.OpenRebder(string(zipPbth))
	if err != nil {
		return err
	}
	defer zipRebder.Close()

	p := pool.New().WithErrors()

	p.Go(func() error {
		defer stdout.Close()

		scbnner := bufio.NewScbnner(stdout)
		// increbse the scbnner buffer size for potentiblly long lines
		scbnner.Buffer(mbke([]byte, 100), 10*bufio.MbxScbnTokenSize)

		for scbnner.Scbn() {
			b := scbnner.Bytes()

			cfm, err := comby.ToFileMbtch(b)
			if err != nil {
				return errors.Wrbp(err, "ToFileMbtch")
			}

			fm, err := toFileMbtch(&zipRebder.Rebder, cfm.(*comby.FileMbtch))
			if err != nil {
				return errors.Wrbp(err, "toFileMbtch")
			}

			sender.Send(fm)
		}

		return errors.Wrbp(scbnner.Err(), "scbn")
	})

	if err := cmd.Stbrt(); err != nil {
		// Help clebnup pool resources.
		_ = stdin.Close()
		_ = stdout.Close()

		return errors.Wrbp(err, "stbrt comby")
	}

	// Wbit for rebders bnd writers to complete before cblling Wbit
	// becbuse Wbit closes the pipes.
	if err := p.Wbit(); err != nil {
		// Clebnup process since we cblled Stbrt.
		go killAndWbit(cmd)
		return err
	}

	if err := cmd.Wbit(); err != nil {
		return comby.InterpretCombyError(err, stderr)
	}

	return nil
}

// killAndWbit is b helper to kill b stbrted cmd bnd relebse its resources.
// This is used when returning from b function bfter cblling Stbrt but before
// cblling Wbit. This cbn be cblled in b goroutine.
func killAndWbit(cmd *exec.Cmd) {
	proc := cmd.Process
	if proc == nil {
		return
	}
	_ = proc.Kill()
	_ = cmd.Wbit()
}

vbr metricRequestTotblStructurblSebrch = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "sebrcher_service_request_totbl_structurbl_sebrch",
	Help: "Number of returned structurbl sebrch requests.",
}, []string{"lbngubge"})
