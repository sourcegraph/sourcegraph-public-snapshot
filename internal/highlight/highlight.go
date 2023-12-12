package highlight

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/binary"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func LoadConfig() {
	client = gosyntect.GetSyntectClient()
}

var (
	client          *gosyntect.Client
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
			Attrs:       []attribute.KeyValue{},
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour { return observation.EmitForHoney },
		})
	})

	return highlightOp
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

	// Format defines the response format of the syntax highlighting request.
	Format gosyntect.HighlightResponseType

	// KeepFinalNewline keeps the final newline of the file content when highlighting.
	// By default we drop the last newline to match behavior of common code hosts
	// that don't render another line at the end of the file.
	KeepFinalNewline bool
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

	scipToHTML(h.code, h.document, addRow, addText, validLines)

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

// identifyError returns true + the problem code if err matches a known error.
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

	logger := log.Scoped("highlight")

	p.Filepath = normalizeFilepath(p.Filepath)

	filetypeQuery := DetectSyntaxHighlightingLanguage(p.Filepath, string(p.Content))

	// Only send tree sitter requests for the languages that we support.
	// TODO: It could be worthwhile to log that this language isn't supported or something
	// like that? Otherwise there is no feedback that this configuration isn't currently working,
	// which is a bit of a confusing situation for the user.
	if !gosyntect.IsTreesitterSupported(filetypeQuery.Language) {
		filetypeQuery.Engine = EngineSyntect
	}

	ctx, errCollector, trace, endObservation := getHighlightOp().WithErrorsAndLogger(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("revision", p.Metadata.Revision),
		attribute.String("repo", p.Metadata.RepoName),
		attribute.String("fileExtension", filepath.Ext(p.Filepath)),
		attribute.String("filepath", p.Filepath),
		attribute.Int("sizeBytes", len(p.Content)),
		attribute.Bool("highlightLongLines", p.HighlightLongLines),
		attribute.Bool("disableTimeout", p.DisableTimeout),
		attribute.Stringer("syntaxEngine", filetypeQuery.Engine),
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
	if binary.IsBinary(p.Content) {
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
	if !p.KeepFinalNewline {
		code = strings.TrimSuffix(code, "\n")
	}

	unhighlightedCode := func(err error, code string) (*HighlightedCode, bool, error) {
		errCollector.Collect(&err)
		plainResponse, tableErr := generatePlainTable(code)
		if tableErr != nil {
			return nil, false, errors.CombineErrors(err, tableErr)
		}
		return plainResponse, true, nil
	}

	if p.Format == gosyntect.FormatHTMLPlaintext {
		return unhighlightedCode(err, code)
	}

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
		LineLengthLimit:  maxLineLength,
		Engine:           getEngineParameter(filetypeQuery.Engine),
	}

	query.Filetype = filetypeQuery.Language

	// Single-program mode: we do not use syntect_server/syntax-highlighter
	//
	// 1. It makes cross-compilation harder (requires a full Rust toolchain for the target, plus
	//    a full C/C++ toolchain for the target.) Complicates macOS code signing.
	// 2. Requires adding a C ABI so we can invoke it via CGO. Or as an external process
	//    complicates distribution and/or requires Docker.
	// 3. syntect_server/syntax-highlighter still uses the absolutely awful http-server-stabilizer
	//    hack to workaround https://github.com/trishume/syntect/issues/202 - and by extension needs
	//    two separate binaries, and separate processes, to function semi-reliably.
	//
	// Instead, in single-program mode we defer to Chroma for syntax highlighting.
	if deploy.IsSingleBinary() {
		document, err := highlightWithChroma(code, p.Filepath)
		if err != nil {
			return unhighlightedCode(err, code)
		}
		if document == nil {
			// Highlighting this language is not supported, so fallback to plain text.
			plainResponse, err := generatePlainTable(code)
			if err != nil {
				return nil, false, err
			}
			return plainResponse, false, nil
		}
		return &HighlightedCode{
			code:     code,
			html:     "",
			document: document,
		}, false, nil
	}

	resp, err := client.Highlight(ctx, query, p.Format)

	if ctx.Err() == context.DeadlineExceeded {
		logger.Warn(
			"syntax highlighting took longer than 3s, this *could* indicate a bug in Sourcegraph",
			log.String("filepath", p.Filepath),
			log.String("filetype", query.Filetype),
			log.String("repo_name", p.Metadata.RepoName),
			log.String("revision", p.Metadata.Revision),
			log.String("snippet", fmt.Sprintf("%q…", firstCharacters(code, 80))),
		)
		trace.AddEvent("syntaxHighlighting", attribute.Bool("timeout", true))
		prometheusStatus = "timeout"

		// Timeout, so render plain table.
		plainResponse, err := generatePlainTable(code)
		if err != nil {
			return nil, false, err
		}
		return plainResponse, true, nil
	} else if err != nil {
		logger.Error(
			"syntax highlighting failed (this is a bug, please report it)",
			log.String("filepath", p.Filepath),
			log.String("filetype", query.Filetype),
			log.String("repo_name", p.Metadata.RepoName),
			log.String("revision", p.Metadata.Revision),
			log.String("snippet", fmt.Sprintf("%q…", firstCharacters(code, 80))),
			log.Error(err),
		)

		if known, problem := identifyError(err); known {
			// A problem that can sometimes be expected has occurred. We will
			// identify such problems through metrics/logs and resolve them on
			// a case-by-case basis.
			trace.AddEvent("TODO Domain Owner", attribute.Bool(problem, true))
			prometheusStatus = problem
		}

		// It is not useful to surface errors in the UI, so fall back to
		// unhighlighted text.
		return unhighlightedCode(err, code)
	}

	// We need to return SCIP data if explicitly requested or if the selected
	// engine is tree sitter.
	if p.Format == gosyntect.FormatJSONSCIP || filetypeQuery.Engine.isTreesitterBased() {
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
//	a/b/c/FOO.TXT → a/b/c/FOO.txt
//	FOO.Sh → FOO.sh
//
// The following are left unmodified, as they already have lowercase extensions:
//
//	a/b/c/FOO.txt
//	a/b/c/Makefile
//	Makefile.am
//	FOO.txt
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
