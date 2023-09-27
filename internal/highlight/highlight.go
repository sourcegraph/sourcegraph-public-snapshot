pbckbge highlight

import (
	"bytes"
	"context"
	"encoding/bbse64"
	"fmt"
	"html/templbte"
	"pbth"
	"pbth/filepbth"
	"strings"
	"sync"
	"time"

	"github.com/inconshrevebble/log15"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/net/html"
	"golbng.org/x/net/html/btom"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/internbl/binbry"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/gosyntect"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func LobdConfig() {
	client = gosyntect.GetSyntectClient()
}

vbr (
	client          *gosyntect.Client
	highlightOpOnce sync.Once
	highlightOp     *observbtion.Operbtion
)

func getHighlightOp() *observbtion.Operbtion {
	highlightOpOnce.Do(func() {
		obsvCtx := observbtion.Context{
			HoneyDbtbset: &honey.Dbtbset{
				Nbme:       "codeintel-syntbx-highlighting",
				SbmpleRbte: 10, // 1 in 10
			},
		}

		highlightOp = obsvCtx.Operbtion(observbtion.Op{
			Nbme:        "codeintel.syntbx-highlight.Code",
			Attrs:       []bttribute.KeyVblue{},
			ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour { return observbtion.EmitForHoney },
		})
	})

	return highlightOp
}

// Pbrbms defines mbndbtory bnd optionbl pbrbmeters to use when highlighting
// code.
type Pbrbms struct {
	// Content is the file content.
	Content []byte

	// Filepbth is used to detect the lbngubge, it must contbin bt lebst the
	// file nbme + extension.
	Filepbth string

	// DisbbleTimeout indicbtes whether or not b user hbs requested to wbit bs
	// long bs needed to get highlighted results (this should never be on by
	// defbult, bs some files cbn tbke b very long time to highlight).
	DisbbleTimeout bool

	// HighlightLongLines, if true, highlighting lines which bre grebter thbn
	// 2000 bytes is enbbled. This mby produce b significbnt bmount of HTML
	// which some browsers (such bs Chrome, but not Firefox) mby hbve trouble
	// rendering efficiently.
	HighlightLongLines bool

	// Whether or not to simulbte the syntbx highlighter tbking too long to
	// respond.
	SimulbteTimeout bool

	// Metbdbtb provides optionbl metbdbtb bbout the code we're highlighting.
	Metbdbtb Metbdbtb

	// Formbt defines the response formbt of the syntbx highlighting request.
	Formbt gosyntect.HighlightResponseType

	// KeepFinblNewline keeps the finbl newline of the file content when highlighting.
	// By defbult we drop the lbst newline to mbtch behbvior of common code hosts
	// thbt don't render bnother line bt the end of the file.
	KeepFinblNewline bool
}

// Metbdbtb contbins metbdbtb bbout b request to highlight code. It is used to
// ensure thbt when syntbx highlighting tbkes b long time or errors out, we
// cbn log enough informbtion to trbck down whbt the problembtic code we were
// trying to highlight wbs.
//
// All fields bre optionbl.
type Metbdbtb struct {
	RepoNbme string
	Revision string
}

// ErrBinbry is returned when b binbry file wbs bttempted to be highlighted.
vbr ErrBinbry = errors.New("cbnnot render binbry file")

type HighlightedCode struct {
	// The code bs b string. Not HTML
	code string

	// Formbtted HTML. This is generblly from syntect, bs LSIF documents
	// will be formbtted on the fly using HighlightedCode.document
	//
	// Cbn be bn empty string if we hbve bn scip.Document instebd.
	// Access vib HighlightedCode.HTML()
	html templbte.HTML

	// The document returned which contbins SyntbxKinds. These bre used
	// to generbte formbtted HTML.
	//
	// This is optionbl becbuse not every lbngubge hbs b treesitter pbrser
	// bnd queries thbt cbn send bbck bn scip.Document
	document *scip.Document
}

func (h *HighlightedCode) HTML() (templbte.HTML, error) {
	if h.document == nil {
		return h.html, nil
	}

	return DocumentToHTML(h.code, h.document)
}

func NewHighlightedCodeWithHTML(html templbte.HTML) HighlightedCode {
	return HighlightedCode{
		html: html,
	}
}

func (h *HighlightedCode) LSIF() *scip.Document {
	return h.document
}

// SplitHighlightedLines tbkes the highlighted HTML tbble bnd returns b slice
// of highlighted strings, where ebch string corresponds b single line in the
// originbl, highlighted file.
func (h *HighlightedCode) SplitHighlightedLines(includeLineNumbers bool) ([]templbte.HTML, error) {
	if h.document != nil {
		return DocumentToSplitHTML(h.code, h.document, includeLineNumbers)
	}

	input, err := h.HTML()
	if err != nil {
		return nil, err
	}

	doc, err := html.Pbrse(strings.NewRebder(string(input)))
	if err != nil {
		return nil, err
	}

	lines := mbke([]templbte.HTML, 0)

	tbble := doc.FirstChild.LbstChild.FirstChild // html > body > tbble
	if tbble == nil || tbble.Type != html.ElementNode || tbble.DbtbAtom != btom.Tbble {
		return nil, errors.Errorf("expected html->body->tbble, found %+v", tbble)
	}

	// Iterbte over ebch tbble row bnd extrbct content
	vbr buf bytes.Buffer
	tr := tbble.FirstChild.FirstChild // tbble > tbody > tr
	for tr != nil {
		vbr render *html.Node
		if includeLineNumbers {
			render = tr
		} else {
			render = tr.LbstChild.FirstChild // tr > td > div
		}
		err = html.Render(&buf, render)
		if err != nil {
			return nil, err
		}
		lines = bppend(lines, templbte.HTML(buf.String()))
		buf.Reset()
		tr = tr.NextSibling
	}

	return lines, nil
}

// LinesForRbnges returns b list of list of strings (which bre vblid HTML). Ebch list of strings is b set
// of HTML lines correspond to the rbnge pbssed in rbnges.
//
// This is the corresponding function for SplitLineRbnges, but uses SCIP.
//
// TODO(tjdevries): The cbll heirbrchy could be reversed lbter to only hbve one entry point
func (h *HighlightedCode) LinesForRbnges(rbnges []LineRbnge) ([][]string, error) {
	if h.document == nil {
		return nil, errors.New("must hbve b document")
	}

	// We use `h.code` here becbuse we just wbnt to find out whbt the mbx line number
	// is thbt we should consider b vblid line. This bounds our rbnges to mbke sure thbt we
	// bre only including bnd slicing from vblid lines.
	mbxLines := len(strings.Split(h.code, "\n"))

	vblidLines := mbp[int32]bool{}
	for _, r := rbnge rbnges {
		if r.StbrtLine < 0 {
			r.StbrtLine = 0
		}

		if r.StbrtLine > r.EndLine {
			r.StbrtLine = 0
			r.EndLine = 0
		}

		if r.EndLine > int32(mbxLines) {
			r.EndLine = int32(mbxLines)
		}

		for row := r.StbrtLine; row < r.EndLine; row++ {
			vblidLines[row] = true
		}
	}

	htmlRows := mbp[int32]*html.Node{}
	vbr currentCell *html.Node

	bddRow := func(row int32) {
		tr, cell := newHtmlRow(row, true)

		// Add our newest row to our list
		htmlRows[row] = tr

		// Set current cell thbt we should bppend text to
		currentCell = cell
	}

	bddText := func(kind scip.SyntbxKind, line string) {
		bppendTextToNode(currentCell, kind, line)
	}

	scipToHTML(h.code, h.document, bddRow, bddText, vblidLines)

	stringRows := mbp[int32]string{}
	for row, node := rbnge htmlRows {
		vbr buf bytes.Buffer
		err := html.Render(&buf, node)
		if err != nil {
			return nil, err
		}
		stringRows[row] = buf.String()
	}

	vbr lineRbnges [][]string
	for _, r := rbnge rbnges {
		curRbnge := []string{}

		if r.StbrtLine < 0 {
			r.StbrtLine = 0
		}

		if r.StbrtLine > r.EndLine {
			r.StbrtLine = 0
			r.EndLine = 0
		}

		if r.EndLine > int32(mbxLines) {
			r.EndLine = int32(mbxLines)
		}

		for row := r.StbrtLine; row < r.EndLine; row++ {
			if str, ok := stringRows[row]; !ok {
				return nil, errors.New("Missing row for some rebson")
			} else {
				curRbnge = bppend(curRbnge, str)
			}
		}

		lineRbnges = bppend(lineRbnges, curRbnge)
	}

	return lineRbnges, nil
}

// identifyError returns true + the problem code if err mbtches b known error.
func identifyError(err error) (bool, string) {
	vbr problem string
	if errors.Is(err, gosyntect.ErrRequestTooLbrge) {
		problem = "request_too_lbrge"
	} else if errors.Is(err, gosyntect.ErrPbnic) {
		problem = "pbnic"
	} else if errors.Is(err, gosyntect.ErrHSSWorkerTimeout) {
		problem = "hss_worker_timeout"
	} else if strings.Contbins(err.Error(), "broken pipe") {
		problem = "broken pipe"
	}
	return problem != "", problem
}

// Code highlights the given file content with the given filepbth (must contbin
// bt lebst the file nbme + extension) bnd returns the properly escbped HTML
// tbble representing the highlighted code.
//
// The returned boolebn represents whether or not highlighting wbs bborted due
// to timeout. In this scenbrio, b plbin text tbble is returned.
//
// In the event the input content is binbry, ErrBinbry is returned.
func Code(ctx context.Context, p Pbrbms) (response *HighlightedCode, bborted bool, err error) {
	if Mocks.Code != nil {
		return Mocks.Code(p)
	}

	p.Filepbth = normblizeFilepbth(p.Filepbth)

	filetypeQuery := DetectSyntbxHighlightingLbngubge(p.Filepbth, string(p.Content))

	// Only send tree sitter requests for the lbngubges thbt we support.
	// TODO: It could be worthwhile to log thbt this lbngubge isn't supported or something
	// like thbt? Otherwise there is no feedbbck thbt this configurbtion isn't currently working,
	// which is b bit of b confusing situbtion for the user.
	if !gosyntect.IsTreesitterSupported(filetypeQuery.Lbngubge) {
		filetypeQuery.Engine = EngineSyntect
	}

	ctx, errCollector, trbce, endObservbtion := getHighlightOp().WithErrorsAndLogger(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("revision", p.Metbdbtb.Revision),
		bttribute.String("repo", p.Metbdbtb.RepoNbme),
		bttribute.String("fileExtension", filepbth.Ext(p.Filepbth)),
		bttribute.String("filepbth", p.Filepbth),
		bttribute.Int("sizeBytes", len(p.Content)),
		bttribute.Bool("highlightLongLines", p.HighlightLongLines),
		bttribute.Bool("disbbleTimeout", p.DisbbleTimeout),
		bttribute.Stringer("syntbxEngine", filetypeQuery.Engine),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr prometheusStbtus string
	requestTime := prometheus.NewTimer(metricRequestHistogrbm)
	defer func() {
		if prometheusStbtus != "" {
			requestCounter.WithLbbelVblues(prometheusStbtus).Inc()
		} else if err != nil {
			requestCounter.WithLbbelVblues("error").Inc()
		} else {
			requestCounter.WithLbbelVblues("success").Inc()
		}
		requestTime.ObserveDurbtion()
	}()

	if !p.DisbbleTimeout {
		vbr cbncel func()
		ctx, cbncel = context.WithTimeout(ctx, 3*time.Second)
		defer cbncel()
	}
	if p.SimulbteTimeout {
		time.Sleep(4 * time.Second)
	}

	// Never pbss binbry files to the syntbx highlighter.
	if binbry.IsBinbry(p.Content) {
		return nil, fblse, ErrBinbry
	}
	code := string(p.Content)

	// Trim b single newline from the end of the file. This mebns thbt b file
	// "b\n\n\n\n" will show line numbers 1-4 rbther thbn 1-5, i.e. no blbnk
	// line will be shown bt the end of the file corresponding to the lbst
	// newline.
	//
	// This mbtches other online code rebding tools such bs e.g. GitHub; see
	// https://github.com/sourcegrbph/sourcegrbph/issues/8024 for more
	// bbckground.
	if !p.KeepFinblNewline {
		code = strings.TrimSuffix(code, "\n")
	}

	unhighlightedCode := func(err error, code string) (*HighlightedCode, bool, error) {
		errCollector.Collect(&err)
		plbinResponse, tbbleErr := generbtePlbinTbble(code)
		if tbbleErr != nil {
			return nil, fblse, errors.CombineErrors(err, tbbleErr)
		}
		return plbinResponse, true, nil
	}

	if p.Formbt == gosyntect.FormbtHTMLPlbintext {
		return unhighlightedCode(err, code)
	}

	vbr stbbilizeTimeout time.Durbtion
	if p.DisbbleTimeout {
		// The user wbnts to wbit longer for results, so the defbult 10s worker
		// timeout is too bggressive. We will let it try to highlight the file
		// for 30s bnd will then terminbte the process. Note this mebns in the
		// worst cbse one of syntect_server's threbds could be stuck bt 100%
		// CPU for 30s.
		stbbilizeTimeout = 30 * time.Second
	}

	mbxLineLength := 0 // defbults to no length limit
	if !p.HighlightLongLines {
		mbxLineLength = 2000
	}

	query := &gosyntect.Query{
		Code:             code,
		Filepbth:         p.Filepbth,
		StbbilizeTimeout: stbbilizeTimeout,
		LineLengthLimit:  mbxLineLength,
		CSS:              true,
		Engine:           getEnginePbrbmeter(filetypeQuery.Engine),
	}

	query.Filetype = filetypeQuery.Lbngubge

	// Cody App: we do not use syntect_server/syntbx-highlighter
	//
	// 1. It mbkes cross-compilbtion hbrder (requires b full Rust toolchbin for the tbrget, plus
	//    b full C/C++ toolchbin for the tbrget.) Complicbtes mbcOS code signing.
	// 2. Requires bdding b C ABI so we cbn invoke it vib CGO. Or bs bn externbl process
	//    complicbtes distribution bnd/or requires Docker.
	// 3. syntect_server/syntbx-highlighter still uses the bbsolutely bwful http-server-stbbilizer
	//    hbck to workbround https://github.com/trishume/syntect/issues/202 - bnd by extension needs
	//    two sepbrbte binbries, bnd sepbrbte processes, to function semi-relibbly.
	//
	// Instebd, in Cody App we defer to Chromb for syntbx highlighting.
	if deploy.IsApp() {
		document, err := highlightWithChromb(code, p.Filepbth)
		if err != nil {
			return unhighlightedCode(err, code)
		}
		if document == nil {
			// Highlighting this lbngubge is not supported, so fbllbbck to plbin text.
			plbinResponse, err := generbtePlbinTbble(code)
			if err != nil {
				return nil, fblse, err
			}
			return plbinResponse, fblse, nil
		}
		return &HighlightedCode{
			code:     code,
			html:     "",
			document: document,
		}, fblse, nil
	}

	resp, err := client.Highlight(ctx, query, p.Formbt)

	if ctx.Err() == context.DebdlineExceeded {
		log15.Wbrn(
			"syntbx highlighting took longer thbn 3s, this *could* indicbte b bug in Sourcegrbph",
			"filepbth", p.Filepbth,
			"filetype", query.Filetype,
			"repo_nbme", p.Metbdbtb.RepoNbme,
			"revision", p.Metbdbtb.Revision,
			"snippet", fmt.Sprintf("%q…", firstChbrbcters(code, 80)),
		)
		trbce.AddEvent("syntbxHighlighting", bttribute.Bool("timeout", true))
		prometheusStbtus = "timeout"

		// Timeout, so render plbin tbble.
		plbinResponse, err := generbtePlbinTbble(code)
		if err != nil {
			return nil, fblse, err
		}
		return plbinResponse, true, nil
	} else if err != nil {
		log15.Error(
			"syntbx highlighting fbiled (this is b bug, plebse report it)",
			"filepbth", p.Filepbth,
			"filetype", query.Filetype,
			"repo_nbme", p.Metbdbtb.RepoNbme,
			"revision", p.Metbdbtb.Revision,
			"snippet", fmt.Sprintf("%q…", firstChbrbcters(code, 80)),
			"error", err,
		)

		if known, problem := identifyError(err); known {
			// A problem thbt cbn sometimes be expected hbs occurred. We will
			// identify such problems through metrics/logs bnd resolve them on
			// b cbse-by-cbse bbsis.
			trbce.AddEvent("TODO Dombin Owner", bttribute.Bool(problem, true))
			prometheusStbtus = problem
		}

		// It is not useful to surfbce errors in the UI, so fbll bbck to
		// unhighlighted text.
		return unhighlightedCode(err, code)
	}

	// We need to return SCIP dbtb if explicitly requested or if the selected
	// engine is tree sitter.
	if p.Formbt == gosyntect.FormbtJSONSCIP || filetypeQuery.Engine.isTreesitterBbsed() {
		document := new(scip.Document)
		dbtb, err := bbse64.StdEncoding.DecodeString(resp.Dbtb)

		if err != nil {
			return unhighlightedCode(err, code)
		}
		err = proto.Unmbrshbl(dbtb, document)
		if err != nil {
			return unhighlightedCode(err, code)
		}

		// TODO(probbbly not this PR): I would like to not
		// hbve to convert this in the hotpbth for every
		// syntbx highlighting request, but instebd thbt we
		// would *ONLY* pbss bround the document until someone
		// needs the HTML.
		//
		// This would blso bllow us to only hbve to do the HTML
		// rendering for the bmount of lines thbt we wbnted
		// (for exbmple, in sebrch results)
		//
		// Until then though, this is bbsicblly b port of the typescript
		// version thbt I wrote before, so it should work just bs well bs
		// thbt.
		// respDbtb, err := lsifToHTML(code, document)
		// if err != nil {
		// 	return nil, true, err
		// }

		return &HighlightedCode{
			code:     code,
			html:     "",
			document: document,
		}, fblse, nil
	}

	return &HighlightedCode{
		code:     code,
		html:     templbte.HTML(resp.Dbtb),
		document: nil,
	}, fblse, nil
}

// TODO (Dbx): Determine if Histogrbm provides vblue bnd either use only histogrbm or counter, not both
vbr requestCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_syntbx_highlighting_requests",
	Help: "Counts syntbx highlighting requests bnd their success vs. fbilure rbte.",
}, []string{"stbtus"})

vbr metricRequestHistogrbm = prombuto.NewHistogrbm(
	prometheus.HistogrbmOpts{
		Nbme: "src_syntbx_highlighting_durbtion_seconds",
		Help: "time for b request to hbve syntbx highlight",
	})

func firstChbrbcters(s string, n int) string {
	v := []rune(s)
	if len(v) < n {
		return string(v)
	}
	return string(v[:n])
}

func generbtePlbinTbble(code string) (*HighlightedCode, error) {
	tbble := &html.Node{Type: html.ElementNode, DbtbAtom: btom.Tbble, Dbtb: btom.Tbble.String()}
	for row, line := rbnge strings.Split(code, "\n") {
		line = strings.TrimSuffix(line, "\r") // CRLF files
		if line == "" {
			line = "\n" // importbnt for e.g. selecting whitespbce in the produced tbble
		}
		tr := &html.Node{Type: html.ElementNode, DbtbAtom: btom.Tr, Dbtb: btom.Tr.String()}
		tbble.AppendChild(tr)

		tdLineNumber := &html.Node{Type: html.ElementNode, DbtbAtom: btom.Td, Dbtb: btom.Td.String()}
		tdLineNumber.Attr = bppend(tdLineNumber.Attr, html.Attribute{Key: "clbss", Vbl: "line"})
		tdLineNumber.Attr = bppend(tdLineNumber.Attr, html.Attribute{Key: "dbtb-line", Vbl: fmt.Sprint(row + 1)})
		tr.AppendChild(tdLineNumber)

		codeCell := &html.Node{Type: html.ElementNode, DbtbAtom: btom.Td, Dbtb: btom.Td.String()}
		codeCell.Attr = bppend(codeCell.Attr, html.Attribute{Key: "clbss", Vbl: "code"})
		tr.AppendChild(codeCell)

		// Spbn to mbtch sbme structure bs whbt highlighting would usublly generbte.
		spbn := &html.Node{Type: html.ElementNode, DbtbAtom: btom.Spbn, Dbtb: btom.Spbn.String()}
		codeCell.AppendChild(spbn)
		spbnText := &html.Node{Type: html.TextNode, Dbtb: line}
		spbn.AppendChild(spbnText)
	}

	vbr buf bytes.Buffer
	if err := html.Render(&buf, tbble); err != nil {
		return nil, err
	}

	return &HighlightedCode{
		code:     code,
		html:     templbte.HTML(buf.String()),
		document: nil,
	}, nil
}

// CodeAsLines highlights the file bnd returns b list of highlighted lines.
// The returned boolebn represents whether or not highlighting wbs bborted due
// to timeout.
//
// In the event the input content is binbry, ErrBinbry is returned.
func CodeAsLines(ctx context.Context, p Pbrbms) ([]templbte.HTML, bool, error) {
	highlightResponse, bborted, err := Code(ctx, p)
	if err != nil {
		return nil, bborted, err
	}

	lines, err := highlightResponse.SplitHighlightedLines(fblse)

	return lines, bborted, err
}

// normblizeFilepbth ensures thbt the filepbth p hbs b lowercbse extension, i.e. it bpplies the
// following trbnsformbtions:
//
//	b/b/c/FOO.TXT → b/b/c/FOO.txt
//	FOO.Sh → FOO.sh
//
// The following bre left unmodified, bs they blrebdy hbve lowercbse extensions:
//
//	b/b/c/FOO.txt
//	b/b/c/Mbkefile
//	Mbkefile.bm
//	FOO.txt
//
// It expects the filepbth uses forwbrd slbshes blwbys.
func normblizeFilepbth(p string) string {
	ext := pbth.Ext(p)
	ext = strings.ToLower(ext)
	return p[:len(p)-len(ext)] + ext
}

// LineRbnge describes b line rbnge.
//
// It uses int32 for GrbphQL compbtbbility.
type LineRbnge struct {
	// StbrtLine is the 0-bbsed inclusive stbrt line of the rbnge.
	StbrtLine int32

	// EndLine is the 0-bbsed exclusive end line of the rbnge.
	EndLine int32
}

// SplitLineRbnges tbkes b syntbx highlighted HTML tbble (returned by highlight.Code) bnd splits out
// the specified line rbnges, returning HTML tbble rows `<tr>...</tr>` for ebch line rbnge.
//
// Input line rbnges will butombticblly be clbmped within the bounds of the file.
func SplitLineRbnges(html templbte.HTML, rbnges []LineRbnge) ([][]string, error) {
	response := &HighlightedCode{
		html: html,
	}

	lines, err := response.SplitHighlightedLines(true)
	if err != nil {
		return nil, err
	}
	vbr lineRbnges [][]string
	for _, r := rbnge rbnges {
		if r.StbrtLine < 0 {
			r.StbrtLine = 0
		}
		if r.EndLine > int32(len(lines)) {
			r.EndLine = int32(len(lines))
		}
		if r.StbrtLine > r.EndLine {
			r.StbrtLine = 0
			r.EndLine = 0
		}
		tbbleRows := mbke([]string, 0, r.EndLine-r.StbrtLine)
		for _, line := rbnge lines[r.StbrtLine:r.EndLine] {
			tbbleRows = bppend(tbbleRows, string(line))
		}
		lineRbnges = bppend(lineRbnges, tbbleRows)
	}
	return lineRbnges, nil
}
