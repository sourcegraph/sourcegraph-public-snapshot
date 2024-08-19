package gcplogurl

import (
	"bytes"
	"strconv"
)

// TruncateFrom provides option for truncate field name.
type TruncateFrom string

const (
	// TruncateFromBeginning about field name.
	TruncateFromBeginning TruncateFrom = "beginning"
	// TruncateFromEnd about field name.
	TruncateFromEnd TruncateFrom = "end"
)

// SummaryFields provides configurations for log's summary fields.
type SummaryFields struct {
	Fields       []string
	Truncate     bool
	MaxLen       int
	TruncateFrom TruncateFrom
}

func (sf *SummaryFields) marshalURL(vs values) {
	vs.Del("summaryFields")
	var last string
	for idx, f := range sf.Fields {
		last = f
		if idx != (len(sf.Fields) - 1) {
			vs.Add("summaryFields", escape(f))
		}
	}
	buf := bytes.NewBufferString(escape(last))
	buf.WriteByte(':')
	buf.WriteString(strconv.FormatBool(sf.Truncate))
	buf.WriteByte(':')
	ml := sf.MaxLen
	if ml == 0 {
		ml = 32
	}
	buf.WriteString(strconv.Itoa(ml))
	buf.WriteByte(':')
	tf := sf.TruncateFrom
	if tf == "" {
		tf = TruncateFromEnd
	}
	buf.WriteString(string(tf))
	vs.Add("summaryFields", buf.String())
}
