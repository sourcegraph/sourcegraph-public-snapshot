package syntaxhighlight

import (
	"encoding/json"
	"io"

	"github.com/sourcegraph/annotate"
)

// JSON annotator produces JSON output but no annotation objects when used by annotate.Annotate
type JSONAnnotator struct {
	// If JSON output should be indented
	indent bool
	// Indicates that output was already started
	dirty bool
	// writer to use
	writer io.Writer
}

// Constructs new JSON annotator by providing indentation mode and writer to use
func NewJSONAnnotator(indent bool, writer io.Writer) *JSONAnnotator {
	return &JSONAnnotator{indent: indent, writer: writer}
}

// Initializes annotator, starts output
func (self *JSONAnnotator) Init() error {
	self.dirty = false
	_, err := self.writer.Write([]byte{'['})
	return err
}

// Shuts down annotator, terminates output
func (self *JSONAnnotator) Done() error {
	var err error
	if self.indent {
		_, err = self.writer.Write([]byte{'\n', ']'})
	} else {
		_, err = self.writer.Write([]byte{']'})
	}
	return err
}

// Writes next token as JSON object
func (self *JSONAnnotator) Annotate(token Token) (*annotate.Annotation, error) {
	var err error
	if self.indent {
		if !self.dirty {
			self.dirty = true
			_, err = self.writer.Write([]byte{'\n', ' ', ' '})
		} else {
			_, err = self.writer.Write([]byte{',', '\n', ' ', ' '})
		}
	} else {
		if self.dirty {
			_, err = self.writer.Write([]byte{','})
		}
	}
	var val []byte
	if self.indent {
		val, err = json.MarshalIndent(token, `  `, `  `)
	} else {
		val, err = json.Marshal(token)
	}
	if err != nil {
		return nil, err
	}
	_, err = self.writer.Write(val)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
