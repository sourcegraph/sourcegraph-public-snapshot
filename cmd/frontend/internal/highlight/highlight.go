package highlight

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/inconshreveable/log15"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/honey"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	syntectServer = env.Get("SRC_SYNTECT_SERVER", "http://syntect-server:9238", "syntect_server HTTP(s) address")
	client        *gosyntect.Client
)

var (
	highlightOpOnce sync.Once
	highlightOp     *observation.Operation
)

func getHighlightOp() *observation.Operation {
	highlightOpOnce.Do(func() {
		obsvCtx := observation.Context{
			HoneyDataset: &honey.Dataset{
				Name:       "codeintel-syntax-highlighting",
				SampleRate: 10, // 1 in 10
			},
		}

		highlightOp = obsvCtx.Operation(observation.Op{
			Name:        "codeintel.syntax-highlight.Code",
			LogFields:   []otlog.Field{},
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour { return observation.EmitForHoney },
		})
	})

	return highlightOp
}

func init() {
	client = gosyntect.New(syntectServer)
}

// IsBinary is a helper to tell if the content of a file is binary or not.
// TODO(tjdevries): This doesn't make sense to be here, IMO
func IsBinary(content []byte) bool {
	// We first check if the file is valid UTF8, since we always consider that
	// to be non-binary.
	//
	// Secondly, if the file is not valid UTF8, we check if the detected HTTP
	// content type is text, which covers a whole slew of other non-UTF8 text
	// encodings for us.
	return !utf8.Valid(content) && !strings.HasPrefix(http.DetectContentType(content), "text/")
}

// Params defines mandatory and optional parameters to use when highlighting
// code.
type Params struct {
	// Content is the file content.
	Content []byte

	// Filepath is used to detect the language, it must contain at least the
	// file name + extension.
	Filepath string

	// DisableTimeout indicates whether or not a user has requested to wait as
	// long as needed to get highlighted results (this should never be on by
	// default, as some files can take a very long time to highlight).
	DisableTimeout bool

	// HighlightLongLines, if true, highlighting lines which are greater than
	// 2000 bytes is enabled. This may produce a significant amount of HTML
	// which some browsers (such as Chrome, but not Firefox) may have trouble
	// rendering efficiently.
	HighlightLongLines bool

	// Whether or not to simulate the syntax highlighter taking too long to
	// respond.
	SimulateTimeout bool

	// Metadata provides optional metadata about the code we're highlighting.
	Metadata Metadata
}

// Metadata contains metadata about a request to highlight code. It is used to
// ensure that when syntax highlighting takes a long time or errors out, we
// can log enough information to track down what the problematic code we were
// trying to highlight was.
//
// All fields are optional.
type Metadata struct {
	RepoName string
	Revision string
}

// ErrBinary is returned when a binary file was attempted to be highlighted.
var ErrBinary = errors.New("cannot render binary file")

type HighlightedCode struct {
	// The code as a string. Not HTML
	code string

	// Formatted HTML. This is generally from syntect, as LSIF documents
	// will be formatted on the fly using HighlightedCode.document
	//
	// Can be an empty string if we have an scip.Document instead.
	// Access via HighlightedCode.HTML()
	html template.HTML

	// The document returned which contains SyntaxKinds. These are used
	// to generate formatted HTML.
	//
	// This is optional because not every language has a treesitter parser
	// and queries that can send back an scip.Document
	document *scip.Document
}

func (h *HighlightedCode) HTML() (template.HTML, error) {
	if h.document == nil {
		return h.html, nil
	}

	return DocumentToHTML(h.code, h.document)
}

func NewHighlightedCodeWithHTML(html template.HTML) HighlightedCode {
	return HighlightedCode{
		html: html,
	}
}

func (h *HighlightedCode) LSIF() *scip.Document {
	return h.document
}

// SplitHighlightedLines takes the highlighted HTML table and returns a slice
// of highlighted strings, where each string corresponds a single line in the
// original, highlighted file.
func (h *HighlightedCode) SplitHighlightedLines(includeLineNumbers bool) ([]template.HTML, error) {
	if h.document != nil {
		return DocumentToSplitHTML(h.code, h.document, includeLineNumbers)
	}

	input, err := h.HTML()
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(input)))
	if err != nil {
		return nil, err
	}

	lines := make([]template.HTML, 0)

	table := doc.FirstChild.LastChild.FirstChild // html > body > table
	if table == nil || table.Type != html.ElementNode || table.DataAtom != atom.Table {
		return nil, errors.Errorf("expected html->body->table, found %+v", table)
	}

	// Iterate over each table row and extract content
	var buf bytes.Buffer
	tr := table.FirstChild.FirstChild // table > tbody > tr
	for tr != nil {
		var render *html.Node
		if includeLineNumbers {
			render = tr
		} else {
			render = tr.LastChild.FirstChild // tr > td > div
		}
		err = html.Render(&buf, render)
		if err != nil {
			return nil, err
		}
		lines = append(lines, template.HTML(buf.String()))
		buf.Reset()
		tr = tr.NextSibling
	}

	return lines, nil
}

// LinesForRanges returns a list of list of strings (which are valid HTML). Each list of strings is a set
// of HTML lines correspond to the range passed in ranges.
//
// This is the corresponding function for SplitLineRanges, but uses SCIP.
//
// TODO(tjdevries): The call heirarchy could be reversed later to only have one entry point
func (h *HighlightedCode) LinesForRanges(ranges []LineRange) ([][]string, error) {
	if h.document == nil {
		return nil, errors.New("must have a document")
	}

	// We use `h.code` here because we just want to find out what the max line number
	// is that we should consider a valid line. This bounds our ranges to make sure that we
	// are only including and slicing from valid lines.
	maxLines := len(strings.Split(h.code, "\n"))

	validLines := map[int32]bool{}
	for _, r := range ranges {
		if r.StartLine < 0 {
			r.StartLine = 0
		}

		if r.StartLine > r.EndLine {
			r.StartLine = 0
			r.EndLine = 0
		}

		if r.EndLine > int32(maxLines) {
			r.EndLine = int32(maxLines)
		}

		for row := r.StartLine; row < r.EndLine; row++ {
			validLines[row] = true
		}
	}

	htmlRows := map[int32]*html.Node{}
	var currentCell *html.Node

	addRow := func(row int32) {
		tr, cell := newHtmlRow(row, true)

		// Add our newest row to our list
		htmlRows[row] = tr

		// Set current cell that we should append text to
		currentCell = cell
	}

	addText := func(kind scip.SyntaxKind, line string) {
		appendTextToNode(currentCell, kind, line)
	}

	lsifToHTML(h.code, h.document, addRow, addText, validLines)

	stringRows := map[int32]string{}
	for row, node := range htmlRows {
		var buf bytes.Buffer
		err := html.Render(&buf, node)
		if err != nil {
			return nil, err
		}
		stringRows[row] = buf.String()
	}

	var lineRanges [][]string
	for _, r := range ranges {
		curRange := []string{}

		if r.StartLine < 0 {
			r.StartLine = 0
		}

		if r.StartLine > r.EndLine {
			r.StartLine = 0
			r.EndLine = 0
		}

		if r.EndLine > int32(maxLines) {
			r.EndLine = int32(maxLines)
		}

		for row := r.StartLine; row < r.EndLine; row++ {
			if str, ok := stringRows[row]; !ok {
				return nil, errors.New("Missing row for some reason")
			} else {
				curRange = append(curRange, str)
			}
		}

		lineRanges = append(lineRanges, curRange)
	}

	return lineRanges, nil
}

/// identifyError returns true + the problem code if err matches a known error.
func identifyError(err error) (bool, string) {
	var problem string
	if errors.Is(err, gosyntect.ErrRequestTooLarge) {
		problem = "request_too_large"
	} else if errors.Is(err, gosyntect.ErrPanic) {
		problem = "panic"
	} else if errors.Is(err, gosyntect.ErrHSSWorkerTimeout) {
		problem = "hss_worker_timeout"
	} else if strings.Contains(err.Error(), "broken pipe") {
		problem = "broken pipe"
	}
	return problem != "", problem
}

// Code highlights the given file content with the given filepath (must contain
// at least the file name + extension) and returns the properly escaped HTML
// table representing the highlighted code.
//
// The returned boolean represents whether or not highlighting was aborted due
// to timeout. In this scenario, a plain text table is returned.
//
// In the event the input content is binary, ErrBinary is returned.
func Code(ctx context.Context, p Params) (response *HighlightedCode, aborted bool, err error) {
	if Mocks.Code != nil {
		return Mocks.Code(p)
	}

	p.Filepath = normalizeFilepath(p.Filepath)

	filetypeQuery := DetectSyntaxHighlightingLanguage(p.Filepath, string(p.Content))

	// Only send tree sitter requests for the languages that we support.
	// TODO: It could be worthwhile to log that this language isn't supported or something
	// like that? Otherwise there is no feedback that this configuration isn't currently working,
	// which is a bit of a confusing situation for the user.
	if !client.IsTreesitterSupported(filetypeQuery.Language) {
		filetypeQuery.Engine = EngineSyntect
	}

	ctx, errCollector, trace, endObservation := getHighlightOp().WithErrorsAndLogger(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("revision", p.Metadata.Revision),
		otlog.String("repo", p.Metadata.RepoName),
		otlog.String("fileExtension", filepath.Ext(p.Filepath)),
		otlog.String("filepath", p.Filepath),
		otlog.Int("sizeBytes", len(p.Content)),
		otlog.Bool("highlightLongLines", p.HighlightLongLines),
		otlog.Bool("disableTimeout", p.DisableTimeout),
		otlog.String("syntaxEngine", engineToDisplay[filetypeQuery.Engine]),
	}})
	defer endObservation(1, observation.Args{})

	var prometheusStatus string
	requestTime := prometheus.NewTimer(metricRequestHistogram)
	defer func() {
		if prometheusStatus != "" {
			requestCounter.WithLabelValues(prometheusStatus).Inc()
		} else if err != nil {
			requestCounter.WithLabelValues("error").Inc()
		} else {
			requestCounter.WithLabelValues("success").Inc()
		}
		requestTime.ObserveDuration()
	}()

	if !p.DisableTimeout {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
	}
	if p.SimulateTimeout {
		time.Sleep(4 * time.Second)
	}

	// Never pass binary files to the syntax highlighter.
	if IsBinary(p.Content) {
		return nil, false, ErrBinary
	}
	code := string(p.Content)

	// Trim a single newline from the end of the file. This means that a file
	// "a\n\n\n\n" will show line numbers 1-4 rather than 1-5, i.e. no blank
	// line will be shown at the end of the file corresponding to the last
	// newline.
	//
	// This matches other online code reading tools such as e.g. GitHub; see
	// https://github.com/sourcegraph/sourcegraph/issues/8024 for more
	// background.
	code = strings.TrimSuffix(code, "\n")

	var stabilizeTimeout time.Duration
	if p.DisableTimeout {
		// The user wants to wait longer for results, so the default 10s worker
		// timeout is too aggressive. We will let it try to highlight the file
		// for 30s and will then terminate the process. Note this means in the
		// worst case one of syntect_server's threads could be stuck at 100%
		// CPU for 30s.
		stabilizeTimeout = 30 * time.Second
	}

	maxLineLength := 0 // defaults to no length limit
	if !p.HighlightLongLines {
		maxLineLength = 2000
	}

	query := &gosyntect.Query{
		Code:             code,
		Filepath:         p.Filepath,
		StabilizeTimeout: stabilizeTimeout,
		Tracer:           ot.GetTracer(ctx),
		LineLengthLimit:  maxLineLength,
		CSS:              true,
	}

	// Set the Filetype part of the command if:
	//    1. We are overriding the config, because then we don't want syntect to try and
	//       guess the filetype (but otherwise we want to maintain backwards compat with
	//       whatever we were calculating before)
	//    2. We are using treesitter. Always have syntect use the language provided in that
	//       case to make sure that we have normalized the names of the language by then.
	if filetypeQuery.LanguageOverride || filetypeQuery.Engine == EngineTreeSitter {
		query.Filetype = filetypeQuery.Language
	}

	resp, err := client.Highlight(ctx, query, filetypeQuery.Engine == EngineTreeSitter)

	unhighlightedCode := func(err error, code string) (*HighlightedCode, bool, error) {
		errCollector.Collect(&err)
		plainResponse, tableErr := generatePlainTable(code)
		if tableErr != nil {
			return nil, false, errors.CombineErrors(err, tableErr)
		}
		return plainResponse, true, nil
	}

	if ctx.Err() == context.DeadlineExceeded {
		log15.Warn(
			"syntax highlighting took longer than 3s, this *could* indicate a bug in Sourcegraph",
			"filepath", p.Filepath,
			"filetype", query.Filetype,
			"repo_name", p.Metadata.RepoName,
			"revision", p.Metadata.Revision,
			"snippet", fmt.Sprintf("%q…", firstCharacters(code, 80)),
		)
		trace.Log(otlog.Bool("timeout", true))
		prometheusStatus = "timeout"

		// Timeout, so render plain table.
		plainResponse, err := generatePlainTable(code)
		if err != nil {
			return nil, false, err
		}
		return plainResponse, true, nil
	} else if err != nil {
		log15.Error(
			"syntax highlighting failed (this is a bug, please report it)",
			"filepath", p.Filepath,
			"filetype", query.Filetype,
			"repo_name", p.Metadata.RepoName,
			"revision", p.Metadata.Revision,
			"snippet", fmt.Sprintf("%q…", firstCharacters(code, 80)),
			"error", err,
		)

		if known, problem := identifyError(err); known {
			// A problem that can sometimes be expected has occurred. We will
			// identify such problems through metrics/logs and resolve them on
			// a case-by-case basis.
			trace.Log(otlog.Bool(problem, true))
			prometheusStatus = problem
		}

		// It is not useful to surface errors in the UI, so fall back to
		// unhighlighted text.
		return unhighlightedCode(err, code)
	}

	if filetypeQuery.Engine == EngineTreeSitter {
		document := new(scip.Document)
		data, err := base64.StdEncoding.DecodeString(resp.Data)

		if err != nil {
			return unhighlightedCode(err, code)
		}
		err = proto.Unmarshal(data, document)
		if err != nil {
			return unhighlightedCode(err, code)
		}

		// TODO(probably not this PR): I would like to not
		// have to convert this in the hotpath for every
		// syntax highlighting request, but instead that we
		// would *ONLY* pass around the document until someone
		// needs the HTML.
		//
		// This would also allow us to only have to do the HTML
		// rendering for the amount of lines that we wanted
		// (for example, in search results)
		//
		// Until then though, this is basically a port of the typescript
		// version that I wrote before, so it should work just as well as
		// that.
		// respData, err := lsifToHTML(code, document)
		// if err != nil {
		// 	return nil, true, err
		// }

		return &HighlightedCode{
			code:     code,
			html:     "",
			document: document,
		}, false, nil
	}

	return &HighlightedCode{
		code:     code,
		html:     template.HTML(resp.Data),
		document: nil,
	}, false, nil
}

// TODO (Dax): Determine if Histogram provides value and either use only histogram or counter, not both
var requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_syntax_highlighting_requests",
	Help: "Counts syntax highlighting requests and their success vs. failure rate.",
}, []string{"status"})

var metricRequestHistogram = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name: "src_syntax_highlighting_duration_seconds",
		Help: "time for a request to have syntax highlight",
	})

func firstCharacters(s string, n int) string {
	v := []rune(s)
	if len(v) < n {
		return string(v)
	}
	return string(v[:n])
}

func generatePlainTable(code string) (*HighlightedCode, error) {
	table := &html.Node{Type: html.ElementNode, DataAtom: atom.Table, Data: atom.Table.String()}
	for row, line := range strings.Split(code, "\n") {
		line = strings.TrimSuffix(line, "\r") // CRLF files
		if line == "" {
			line = "\n" // important for e.g. selecting whitespace in the produced table
		}
		tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}
		table.AppendChild(tr)

		tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
		tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "class", Val: "line"})
		tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "data-line", Val: fmt.Sprint(row + 1)})
		tr.AppendChild(tdLineNumber)

		codeCell := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
		codeCell.Attr = append(codeCell.Attr, html.Attribute{Key: "class", Val: "code"})
		tr.AppendChild(codeCell)

		// Span to match same structure as what highlighting would usually generate.
		span := &html.Node{Type: html.ElementNode, DataAtom: atom.Span, Data: atom.Span.String()}
		codeCell.AppendChild(span)
		spanText := &html.Node{Type: html.TextNode, Data: line}
		span.AppendChild(spanText)
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, table); err != nil {
		return nil, err
	}

	return &HighlightedCode{
		code:     code,
		html:     template.HTML(buf.String()),
		document: nil,
	}, nil
}

// CodeAsLines highlights the file and returns a list of highlighted lines.
// The returned boolean represents whether or not highlighting was aborted due
// to timeout.
//
// In the event the input content is binary, ErrBinary is returned.
func CodeAsLines(ctx context.Context, p Params) ([]template.HTML, bool, error) {
	highlightResponse, aborted, err := Code(ctx, p)
	if err != nil {
		return nil, aborted, err
	}

	lines, err := highlightResponse.SplitHighlightedLines(false)

	return lines, aborted, err
}

// normalizeFilepath ensures that the filepath p has a lowercase extension, i.e. it applies the
// following transformations:
//
// 	a/b/c/FOO.TXT → a/b/c/FOO.txt
// 	FOO.Sh → FOO.sh
//
// The following are left unmodified, as they already have lowercase extensions:
//
// 	a/b/c/FOO.txt
// 	a/b/c/Makefile
// 	Makefile.am
// 	FOO.txt
//
// It expects the filepath uses forward slashes always.
func normalizeFilepath(p string) string {
	ext := path.Ext(p)
	ext = strings.ToLower(ext)
	return p[:len(p)-len(ext)] + ext
}

// LineRange describes a line range.
//
// It uses int32 for GraphQL compatability.
type LineRange struct {
	// StartLine is the 0-based inclusive start line of the range.
	StartLine int32

	// EndLine is the 0-based exclusive end line of the range.
	EndLine int32
}

// SplitLineRanges takes a syntax highlighted HTML table (returned by highlight.Code) and splits out
// the specified line ranges, returning HTML table rows `<tr>...</tr>` for each line range.
//
// Input line ranges will automatically be clamped within the bounds of the file.
func SplitLineRanges(html template.HTML, ranges []LineRange) ([][]string, error) {
	response := &HighlightedCode{
		html: html,
	}

	lines, err := response.SplitHighlightedLines(true)
	if err != nil {
		return nil, err
	}
	var lineRanges [][]string
	for _, r := range ranges {
		if r.StartLine < 0 {
			r.StartLine = 0
		}
		if r.EndLine > int32(len(lines)) {
			r.EndLine = int32(len(lines))
		}
		if r.StartLine > r.EndLine {
			r.StartLine = 0
			r.EndLine = 0
		}
		tableRows := make([]string, 0, r.EndLine-r.StartLine)
		for _, line := range lines[r.StartLine:r.EndLine] {
			tableRows = append(tableRows, string(line))
		}
		lineRanges = append(lineRanges, tableRows)
	}
	return lineRanges, nil
}
