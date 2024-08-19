// Package v4 provides a decoder for Jupyter Notebooks v4.0 and later minor versions.
//
// It implements the IPython Notebook v4.0 JSON Schema. Other minor versions can be decoded using the same,
// as the differences do not affect how the notebook is rendered.
//
// [IPython Notebook v4.0 JSON Schema]: https://github.com/jupyter/nbformat/blob/main/nbformat/v4/nbformat.v4.0.schema.json
package v4

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bevzzz/nb/decode"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

func init() {
	d := new(decoder)
	decode.RegisterDecoder(schema.Version{Major: 4, Minor: 5}, d)
	decode.RegisterDecoder(schema.Version{Major: 4, Minor: 4}, d)
	decode.RegisterDecoder(schema.Version{Major: 4, Minor: 3}, d)
	decode.RegisterDecoder(schema.Version{Major: 4, Minor: 2}, d)
	decode.RegisterDecoder(schema.Version{Major: 4, Minor: 1}, d)
	decode.RegisterDecoder(schema.Version{Major: 4, Minor: 0}, d)
}

// decoder decodes cell contents and metadata for nbformat v4.0.
type decoder struct{}

var _ decode.Decoder = (*decoder)(nil)

func (d *decoder) ExtractCells(data []byte) ([]json.RawMessage, error) {
	var raw struct {
		Cells []json.RawMessage `json:"cells"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw.Cells, nil
}

func (d *decoder) DecodeMeta(data []byte) (schema.NotebookMetadata, error) {
	var nm NotebookMetadata
	if err := json.Unmarshal(data, &nm); err != nil {
		return nil, err
	}
	return &nm, nil
}

func (d *decoder) DecodeCell(m map[string]interface{}, data []byte, meta schema.NotebookMetadata) (schema.Cell, error) {
	var ct interface{}
	var c schema.Cell
	switch ct = m["cell_type"]; ct {
	case "markdown":
		c = &Markdown{}
	case "raw":
		c = &Raw{}
	case "code":
		c = &Code{Lang: meta.Language()}
	default:
		return nil, fmt.Errorf("unknown cell type %q", ct)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%s: %w", ct, err)
	}
	return c, nil
}

type NotebookMetadata struct {
	Lang struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"language_info"`
}

var _ schema.NotebookMetadata = (*NotebookMetadata)(nil)

func (nm *NotebookMetadata) Language() string {
	return nm.Lang.Name
}

// Markdown defines the schema for a "markdown" cell.
type Markdown struct {
	common.Markdown
	Att Attachments `json:"attachments,omitempty"`
}

var _ schema.HasAttachments = (*Markdown)(nil)

func (md *Markdown) Attachments() schema.Attachments {
	return md.Att
}

// Raw defines the schema for a "raw" cell.
type Raw struct {
	common.Raw
	Att Attachments `json:"attachments,omitempty"`
}

var _ schema.HasAttachments = (*Raw)(nil)

func (raw *Raw) Attachments() schema.Attachments {
	return raw.Att
}

// Attachments store mime-bundles keyed by filename.
type Attachments map[string]MimeBundle

var _ schema.Attachments = new(Attachments)

func (att Attachments) MimeBundle(filename string) schema.MimeBundle {
	mb, ok := att[filename]
	if !ok {
		return nil
	}
	return mb
}

// Code defines the schema for a "code" cell.
type Code struct {
	Source        common.MultilineString `json:"source"`
	TimesExecuted int                    `json:"execution_count"`
	Out           []Output               `json:"outputs"`
	Lang          string                 `json:"-"`
}

var _ schema.CodeCell = (*Code)(nil)
var _ schema.Outputter = (*Code)(nil)

func (code *Code) Type() schema.CellType {
	return schema.Code
}

// FIXME: return correct mime type (add a function to common)
func (code *Code) MimeType() string {
	return "application/x-python"
}

func (code *Code) Text() []byte {
	return code.Source.Text()
}

func (code *Code) Language() string {
	return code.Lang
}

func (code *Code) ExecutionCount() int {
	return code.TimesExecuted
}

func (code *Code) Outputs() (cells []schema.Cell) {
	for i := range code.Out {
		cells = append(cells, code.Out[i].cell)
	}
	return
}

// Outputs unmarshals cell outputs into schema.Cell based on their type.
type Output struct {
	cell schema.Cell
}

func (out *Output) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("code outputs: %w", err)
	}

	var t interface{}
	var c schema.Cell
	switch t = v["output_type"]; t {
	case "stream":
		c = &StreamOutput{}
	case "display_data":
		c = &DisplayDataOutput{}
	case "execute_result":
		c = &ExecuteResultOutput{}
	case "error":
		c = &ErrorOutput{}
	default:
		return fmt.Errorf("unknown output type %q", t)
	}

	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("%q output: %w", t, err)
	}
	out.cell = c
	return nil
}

// StreamOutput is a plain, text-based output of the executed code.
// Depending on the stream "target", Type() can report "text/plain" (stdout) or "error" (stderr).
// The output is often decorated with ANSI-color sequences, which should be handled separately.
type StreamOutput struct {
	// Target can be stdout or stderr.
	Target string                 `json:"name"`
	Source common.MultilineString `json:"text"`
}

var _ schema.Cell = (*StreamOutput)(nil)

func (stream *StreamOutput) Type() schema.CellType {
	return schema.Stream
}

func (stream *StreamOutput) MimeType() string {
	switch stream.Target {
	case "stdout":
		return common.Stdout
	case "stderr":
		return common.Stderr
	}
	return common.PlainText
}

func (stream *StreamOutput) Text() []byte {
	return stream.Source.Text()
}

// DisplayDataOutput are rich-format outputs generated by running the code in the parent cell.
type DisplayDataOutput struct {
	MimeBundle `json:"data"`
	Metadata   map[string]interface{} `json:"metadata"`
}

var _ schema.Cell = (*DisplayDataOutput)(nil)

func (dd *DisplayDataOutput) Type() schema.CellType {
	return schema.DisplayData
}

// MimeBundle contains rich output data keyed by mime-type.
type MimeBundle map[string]interface{}

var _ schema.MimeBundle = (*MimeBundle)(nil)

// MimeType returns the richer of the mime-types present in the bundle,
// and falls back to "text/plain" otherwise.
func (mb MimeBundle) MimeType() string {
	for mime := range mb {
		if mime != common.PlainText {
			return mime
		}
	}
	return common.PlainText
}

// Text returns data with the richer mime-type.
func (mb MimeBundle) Text() []byte {
	return mb.Data(mb.MimeType())
}

// Data returns mime-type-specific content if present and a nil slice otherwise.
func (mb MimeBundle) Data(mime string) []byte {
	if txt, ok := mb[mime]; ok {

		switch v := txt.(type) {
		case []byte:
			return v
		case string:
			return []byte(v)
		// case []string: TODO: handle MultilineString case
		case map[string]interface{}:
			// TODO(optimization): see if there's a way to keep this as raw bytes during unmarshaling to doing the work twice.
			if b, err := json.Marshal(txt); err == nil {
				return b
			}
			return nil
		}
	}
	return nil
}

// PlainText returns data for "text/plain" mime-type and a nil slice otherwise.
func (mb MimeBundle) PlainText() []byte {
	return mb.Data(common.PlainText)
}

// ExecuteResultOutput is the result of executing the code in the cell.
// Its contents are identical to those of DisplayDataOutput with the addition of the execution count.
type ExecuteResultOutput struct {
	DisplayDataOutput
	TimesExecuted int `json:"execution_count"`
}

var _ schema.Cell = (*ExecuteResultOutput)(nil)
var _ schema.ExecutionCounter = (*ExecuteResultOutput)(nil)

func (ex *ExecuteResultOutput) Type() schema.CellType {
	return schema.ExecuteResult
}

func (ex *ExecuteResultOutput) ExecutionCount() int {
	return ex.TimesExecuted
}

// ErrorOutput stores the output of a failed code execution.
type ErrorOutput struct {
	ExceptionName  string   `json:"ename"`
	ExceptionValue string   `json:"evalue"`
	Traceback      []string `json:"traceback"`
}

var _ schema.Cell = (*ErrorOutput)(nil)

func (err *ErrorOutput) Type() schema.CellType {
	return schema.Error
}

func (err *ErrorOutput) MimeType() string {
	return common.Stderr
}

func (err *ErrorOutput) Text() (txt []byte) {
	s := strings.Join(err.Traceback, "\n")
	return []byte(s)
}
