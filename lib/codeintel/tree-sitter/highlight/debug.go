package highlight

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	tree_sitter "github.com/smacker/go-tree-sitter"
)

//----------------------------------------------------------------------------
// Debugging helpers

type SexprFormatter struct {
	indent int
	input  []byte
	output strings.Builder
	spaces []byte
}

func NewSexprFormatter(input []byte) SexprFormatter {
	return SexprFormatter{
		0, input, strings.Builder{}, []byte("        "),
	}
}

func (f *SexprFormatter) Format(node *tree_sitter.Node) string {
	iter := NewAllOrderIterator(node, f)
	if _, err := iter.VisitTree(); err != nil {
		panic(errors.Wrapf(err, "failed to format parse tree"))
	}
	return f.output.String()
}

func (f *SexprFormatter) writeLeadingIndentation() {
	if f.indent == 0 {
		return
	}
	if f.indent > len(f.spaces) {
		f.spaces = bytes.Repeat(f.spaces, (f.indent/len(f.spaces))+1)
		if f.indent > len(f.spaces) {
			panic("spaces buffer should've grown to cover indentation")
		}
	}
	if _, err := io.CopyN(&f.output, bytes.NewReader(f.spaces), int64(f.indent)); err != nil {
		panic(fmt.Sprintf("failed to copy spaces from buffer, copy_size=%d", f.indent))
	}
}

func (f *SexprFormatter) startVisit(node *tree_sitter.Node) error {
	var type_ string
	if node.IsNull() {
		type_ = "NULL"
	} else if node.IsMissing() {
		type_ = "MISSING"
	} else if !node.IsNamed() {
		type_ = "_anon"
	} else {
		type_ = node.Type()
	}
	f.output.WriteByte('\n')
	f.writeLeadingIndentation()
	f.output.WriteString(fmt.Sprintf("(%s [%d:%d %d:%d]", type_, node.StartPoint().Row+1,
		node.StartPoint().Column+1, node.EndPoint().Row+1, node.EndPoint().Column+1))
	if node.NamedChildCount() == 0 {
		f.output.WriteString(fmt.Sprintf(" text=%#v", node.Content(f.input)))
	}
	f.indent += 2
	return nil
}

func (f *SexprFormatter) afterVisitingChild(parentNode *tree_sitter.Node, visitedChildIndex int) error {
	return nil
}

func (f *SexprFormatter) endVisit(n *tree_sitter.Node) error {
	f.indent -= 2
	f.output.WriteByte(')')
	return nil
}

func (f *SexprFormatter) visitAnonymous(n *tree_sitter.Node) error {
	//TODO implement me
	if err := f.startVisit(n); err != nil {
		return err
	}
	return f.endVisit(n)
}

var _ TreeVisitor = &SexprFormatter{}
