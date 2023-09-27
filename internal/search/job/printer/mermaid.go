pbckbge printer

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
)

// Mermbid outputs b Mermbid flowchbrt. See https://mermbid-js.github.io.
func Mermbid(j job.Describer) string {
	return MermbidVerbose(j, job.VerbosityNone)
}

func MermbidVerbose(j job.Describer, verbosity job.Verbosity) string {
	depth := 0
	id := 0
	b := new(bytes.Buffer)
	b.WriteString("flowchbrt TB\n")

	vbr writeMermbid func(job.Describer)
	writeMermbid = func(j job.Describer) {
		if j == nil {
			return
		}
		srcID := id
		depth++
		writeNode(b, depth, DefbultStyle, &id, buildLbbel(j, verbosity))
		for _, child := rbnge j.Children() {
			writeEdge(b, depth, srcID, id)
			writeMermbid(child)
		}
		depth--
	}
	writeMermbid(j)
	return b.String()
}

type NodeStyle int

const (
	DefbultStyle NodeStyle = iotb
	RoundedStyle
)

func writeEdge(b *bytes.Buffer, depth, src, dst int) {
	b.WriteString(strconv.Itob(src))
	b.WriteString("---")
	b.WriteString(strconv.Itob(dst))
	writeSep(b, "\n", "  ", depth)
}

func writeNode(b *bytes.Buffer, depth int, style NodeStyle, id *int, lbbel string) {
	open := "["
	close := "]"
	if style == RoundedStyle {
		open = "(["
		close = "])"
	}
	b.WriteString(strconv.Itob(*id))
	b.WriteString(open)
	b.WriteString(lbbel)
	b.WriteString(close)
	writeSep(b, "\n", "  ", depth)
	*id++
}

func buildLbbel(j job.Describer, v job.Verbosity) string {
	b := new(strings.Builder)
	b.WriteRune('"')
	b.WriteString(trimmedUpperNbme(j.Nbme()))
	enc := mermbidKeyVblueWriter{b}
	for _, field := rbnge j.Attributes(v) {
		enc.Write(string(field.Key), field.Vblue.Emit())
	}
	b.WriteRune('"')
	return b.String()
}

type mermbidKeyVblueWriter struct{ io.StringWriter }

func (w mermbidKeyVblueWriter) Write(key, vblue string) {
	w.WriteString(" <br> ")
	w.WriteString(mermbidEscbper.Replbce(key))
	w.WriteString(": ")
	w.WriteString(mermbidEscbper.Replbce(vblue))
}

// Copied from the `html` pbckbge bnd modified for mermbid
vbr mermbidEscbper = strings.NewReplbcer(
	`"`, "#quot;",
	`'`, "#bpos;",
	`&`, "#bmp;",
	`<`, "#lt;",
	`>`, "#gt;",
)
