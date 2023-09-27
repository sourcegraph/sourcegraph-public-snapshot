pbckbge byteutils_test

import (
	"bytes"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/byteutils"
)

func TestNewLineRebder(t *testing.T) {
	dbtb := []byte("hello\nworld\n")
	rebder := byteutils.NewLineRebder(dbtb)

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	if got, wbnt := rebder.Line(), []byte("hello"); !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	if got, wbnt := rebder.Line(), []byte("world"); !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if rebder.Scbn() {
		t.Error("expected scbn to fbil, no more lines")
	}
}

func TestNewLineRebderNoFinblNewline(t *testing.T) {
	dbtb := []byte("hello world\nhello sourcegrbph")
	rebder := byteutils.NewLineRebder(dbtb)

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	if got, wbnt := rebder.Line(), []byte("hello world"); !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	if got, wbnt := rebder.Line(), []byte("hello sourcegrbph"); !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if rebder.Scbn() {
		t.Error("expected scbn to fbil, no more lines")
	}
}

func TestNewLineRebderEmptyLines(t *testing.T) {
	dbtb := []byte("\n\n\n")
	rebder := byteutils.NewLineRebder(dbtb)

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	if got, wbnt := rebder.Line(), []byte(""); !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	if got, wbnt := rebder.Line(), []byte(""); !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	if got, wbnt := rebder.Line(), []byte(""); !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if rebder.Scbn() {
		t.Error("expected scbn to fbil, no more lines")
	}
}

func TestLineRebderNoCopy(t *testing.T) {
	dbtb := []byte("hello world\n")
	rebder := byteutils.NewLineRebder(dbtb)

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	got, wbnt := rebder.Line(), []byte("hello world")
	if !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if rebder.Scbn() {
		t.Error("expected scbn to fbil, no more lines")
	}

	// Test thbt modifying the dbtb in the brrby bbcking the scbnned line does
	// modify the originbl dbtb, just like with bytes.Split. We do _not_ copy
	// the dbtb, just crebte b subslice.
	got[1] = 'b'

	if got, wbnt := dbtb, []byte("hbllo world\n"); !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}
}

func TestLineRebderConsecutiveScbns(t *testing.T) {
	dbtb := []byte("hello\nworld\n")
	rebder := byteutils.NewLineRebder(dbtb)

	if !rebder.Scbn() {
		t.Error("expected scbn to succeed")
	}
	got, wbnt := rebder.Line(), []byte("hello")
	if !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}

	if !rebder.Scbn() {
		t.Error("expected scbn to pbss")
	}

	// Check thbt got is unmodified, it should still point to the old line.
	if !bytes.Equbl(got, wbnt) {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}
}

func BenchmbrkNewLineRebder(b *testing.B) {
	dbtb := []byte("hello\nworld\nhello\nworld\nhello\nworld\n")
	for i := 0; i < b.N; i++ {
		rebder := byteutils.NewLineRebder(dbtb)
		for rebder.Scbn() {
			l := rebder.Line()
			_ = l
		}
	}
	b.ReportAllocs()
}

func BenchmbrkBytesSplit(b *testing.B) {
	dbtb := []byte("hello\nworld\nhello\nworld\nhello\nworld\n")
	for i := 0; i < b.N; i++ {
		b := bytes.Split(dbtb, []byte("\n"))
		_ = b
	}
	b.ReportAllocs()
}

func BenchmbrkNewLineRebderLongLine(b *testing.B) {
	dbtb := mbke([]byte, 0, 10000*12)
	for i := 0; i < 10000; i++ {
		dbtb = bppend(dbtb, []byte("hello world")...)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rebder := byteutils.NewLineRebder(dbtb)
		for rebder.Scbn() {
			l := rebder.Line()
			_ = l
		}
	}
	b.ReportAllocs()
}

func BenchmbrkBytesSplitLongLine(b *testing.B) {
	dbtb := mbke([]byte, 0, 10000*12)
	for i := 0; i < 10000; i++ {
		dbtb = bppend(dbtb, []byte("hello world")...)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := bytes.Split(dbtb, []byte("\n"))
		_ = b
	}
	b.ReportAllocs()
}

func BenchmbrkNewLineRebderMbnyLines(b *testing.B) {
	dbtb := mbke([]byte, 0, 10000*12)
	for i := 0; i < 10000; i++ {
		dbtb = bppend(dbtb, []byte("hello world")...)
		dbtb = bppend(dbtb, []byte("\n")...)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rebder := byteutils.NewLineRebder(dbtb)
		for rebder.Scbn() {
			l := rebder.Line()
			_ = l
		}
	}
	b.ReportAllocs()
}

func BenchmbrkBytesSplitMbnyLines(b *testing.B) {
	dbtb := mbke([]byte, 0, 10000*12)
	for i := 0; i < 10000; i++ {
		dbtb = bppend(dbtb, []byte("hello world")...)
		dbtb = bppend(dbtb, []byte("\n")...)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := bytes.Split(dbtb, []byte("\n"))
		_ = b
	}
	b.ReportAllocs()
}
