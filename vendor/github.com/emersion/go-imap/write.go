package imap

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type flusher interface {
	Flush() error
}

type (
	// A string that will be quoted.
	Quoted string
	// A raw atom.
	Atom string
)

type WriterTo interface {
	WriteTo(w *Writer) error
}

func formatNumber(num uint32) string {
	return strconv.FormatUint(uint64(num), 10)
}

// Convert a string list to a field list.
func FormatStringList(list []string) (fields []interface{}) {
	fields = make([]interface{}, len(list))
	for i, v := range list {
		fields[i] = v
	}
	return
}

// Check if a string is 8-bit clean.
func isAscii(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII || unicode.IsControl(c) {
			return false
		}
	}
	return true
}

// An IMAP writer.
type Writer struct {
	io.Writer

	continues <-chan bool
}

// Helper function to write a string to w.
func (w *Writer) writeString(s string) error {
	_, err := io.WriteString(w.Writer, s)
	return err
}

func (w *Writer) writeCrlf() error {
	if err := w.writeString(crlf); err != nil {
		return err
	}

	return w.Flush()
}

func (w *Writer) writeNumber(num uint32) error {
	return w.writeString(formatNumber(num))
}

func (w *Writer) writeQuoted(s string) error {
	return w.writeString(strconv.Quote(s))
}

func (w *Writer) writeAtom(s string) error {
	return w.writeString(s)
}

func (w *Writer) writeAstring(s string) error {
	if !isAscii(s) {
		// IMAP doesn't allow 8-bit data outside literals
		return w.writeLiteral(bytes.NewBufferString(s))
	}

	if strings.ToUpper(s) == nilAtom || s == "" || strings.ContainsAny(s, atomSpecials) {
		return w.writeQuoted(s)
	}

	return w.writeAtom(s)
}

func (w *Writer) writeDateTime(t time.Time, layout string) error {
	if t.IsZero() {
		return w.writeAtom(nilAtom)
	}
	return w.writeQuoted(t.Format(layout))
}

func (w *Writer) writeFields(fields []interface{}) error {
	for i, field := range fields {
		if i > 0 { // Write separator
			if err := w.writeString(string(sp)); err != nil {
				return err
			}
		}

		if err := w.writeField(field); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) writeList(fields []interface{}) error {
	if err := w.writeString(string(listStart)); err != nil {
		return err
	}

	if err := w.writeFields(fields); err != nil {
		return err
	}

	return w.writeString(string(listEnd))
}

func (w *Writer) writeLiteral(l Literal) error {
	if l == nil {
		return w.writeString(nilAtom)
	}

	header := string(literalStart) + strconv.Itoa(l.Len()) + string(literalEnd) + crlf
	if err := w.writeString(header); err != nil {
		return err
	}

	// If a channel is available, wait for a continuation request before sending data
	if w.continues != nil {
		// Make sure to flush the writer, otherwise we may never receive a continuation request
		if err := w.Flush(); err != nil {
			return err
		}

		if !<-w.continues {
			return fmt.Errorf("imap: cannot send literal: no continuation request received")
		}
	}

	_, err := io.Copy(w, l)
	return err
}

func (w *Writer) writeField(field interface{}) error {
	if field == nil {
		return w.writeAtom(nilAtom)
	}

	switch field := field.(type) {
	case string:
		return w.writeAstring(field)
	case Quoted:
		return w.writeQuoted(string(field))
	case Atom:
		return w.writeAtom(string(field))
	case int:
		return w.writeNumber(uint32(field))
	case uint32:
		return w.writeNumber(field)
	case Literal:
		return w.writeLiteral(field)
	case []interface{}:
		return w.writeList(field)
	case envelopeDateTime:
		return w.writeDateTime(time.Time(field), envelopeDateTimeLayout)
	case searchDate:
		return w.writeDateTime(time.Time(field), searchDateLayout)
	case Date:
		return w.writeDateTime(time.Time(field), DateLayout)
	case DateTime:
		return w.writeDateTime(time.Time(field), DateTimeLayout)
	case time.Time:
		return w.writeDateTime(field, DateTimeLayout)
	case *SeqSet:
		return w.writeString(field.String())
	case *BodySectionName:
		// Can contain spaces - that's why we don't just pass it as a string
		return w.writeString(string(field.FetchItem()))
	}

	return fmt.Errorf("imap: cannot format field: %v", field)
}

func (w *Writer) writeRespCode(code StatusRespCode, args []interface{}) error {
	if err := w.writeString(string(respCodeStart)); err != nil {
		return err
	}

	fields := []interface{}{string(code)}
	fields = append(fields, args...)

	if err := w.writeFields(fields); err != nil {
		return err
	}

	return w.writeString(string(respCodeEnd))
}

func (w *Writer) writeLine(fields ...interface{}) error {
	if err := w.writeFields(fields); err != nil {
		return err
	}

	return w.writeCrlf()
}

func (w *Writer) Flush() error {
	if f, ok := w.Writer.(flusher); ok {
		return f.Flush()
	}
	return nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

func NewClientWriter(w io.Writer, continues <-chan bool) *Writer {
	return &Writer{Writer: w, continues: continues}
}
