package fileset

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func checkPos(t *testing.T, msg string, p, q Position) {
	if p.Filename != q.Filename {
		t.Errorf("%s: expected filename = %q; got %q", msg, q.Filename, p.Filename)
	}
	if p.Offset != q.Offset {
		t.Errorf("%s: expected offset = %d; got %d", msg, q.Offset, p.Offset)
	}
	if p.Line != q.Line {
		t.Errorf("%s: expected line = %d; got %d", msg, q.Line, p.Line)
	}
	if p.Column != q.Column {
		t.Errorf("%s: expected column = %d; got %d", msg, q.Column, p.Column)
	}
}

func TestNoPos(t *testing.T) {
	if NoPos.IsValid() {
		t.Errorf("NoPos should not be valid")
	}
	var fset *FileSet
	checkPos(t, "nil NoPos", fset.Position(NoPos), Position{})
	fset = NewFileSet()
	checkPos(t, "fset NoPos", fset.Position(NoPos), Position{})
}

var tests = []struct {
	filename string
	source   []byte // may be nil
	size     int
	lines    []int
}{
	{"a", []byte{}, 0, []int{}},
	{"b", []byte("01234"), 5, []int{0}},
	{"c", []byte("\n\n\n\n\n\n\n\n\n"), 9, []int{0, 1, 2, 3, 4, 5, 6, 7, 8}},
	{"d", nil, 100, []int{0, 5, 10, 20, 30, 70, 71, 72, 80, 85, 90, 99}},
	{"e", nil, 777, []int{0, 80, 100, 120, 130, 180, 267, 455, 500, 567, 620}},
	{"f", []byte("package p\n\nimport \"fmt\""), 23, []int{0, 10, 11}},
	{"g", []byte("package p\n\nimport \"fmt\"\n"), 24, []int{0, 10, 11}},
	{"h", []byte("package p\n\nimport \"fmt\"\n "), 25, []int{0, 10, 11, 24}},
}

func linecol(lines []int, offs int) (int, int) {
	prevLineOffs := 0
	for line, lineOffs := range lines {
		if offs < lineOffs {
			return line, offs - prevLineOffs + 1
		}
		prevLineOffs = lineOffs
	}
	return len(lines), offs - prevLineOffs + 1
}

func verifyPositions(t *testing.T, fset *FileSet, f *File, lines []int) {
	for offs := 0; offs < f.Size(); offs++ {
		p := f.Pos(offs)
		offs2 := f.Offset(p)
		if offs2 != offs {
			t.Errorf("%s, Offset: expected offset %d; got %d", f.Name(), offs, offs2)
		}
		line, col := linecol(lines, offs)
		msg := fmt.Sprintf("%s (offs = %d, p = %d)", f.Name(), offs, p)
		checkPos(t, msg, f.Position(f.Pos(offs)), Position{f.Name(), offs, line, col})
		checkPos(t, msg, fset.Position(p), Position{f.Name(), offs, line, col})
	}
}

func verifyLineOffsets(t *testing.T, fset *FileSet, f *File, lines []int) {
	for i, offs := range lines {
		offs2 := f.LineOffset(i + 1)
		if offs2 != offs {
			t.Errorf("%s, LineOffset: expected line offset %d; got %d", f.Name(), offs, offs2)
		}

		var eoffs int
		if i == len(lines)-1 {
			eoffs = f.Size()
		} else {
			eoffs = lines[i+1] - 1
		}
		eoffs2 := f.LineEndOffset(i + 1)
		if eoffs2 != eoffs {
			t.Errorf("%s, LineEndOffset: expected line end offset %d; got %d", f.Name(), eoffs, eoffs2)
		}
	}
}

func makeTestSource(size int, lines []int) []byte {
	src := make([]byte, size)
	for _, offs := range lines {
		if offs > 0 {
			src[offs-1] = '\n'
		}
	}
	return src
}

func TestPositions(t *testing.T) {
	const delta = 7 // a non-zero base offset increment
	fset := NewFileSet()
	for _, test := range tests {
		// verify consistency of test case
		if test.source != nil && len(test.source) != test.size {
			t.Errorf("%s: inconsistent test case: expected file size %d; got %d", test.filename, test.size, len(test.source))
		}

		// add file and verify name and size
		f := fset.AddFile(test.filename, fset.Base()+delta, test.size)
		if f.Name() != test.filename {
			t.Errorf("expected filename %q; got %q", test.filename, f.Name())
		}
		if f.Size() != test.size {
			t.Errorf("%s: expected file size %d; got %d", f.Name(), test.size, f.Size())
		}
		if fset.File(f.Pos(0)) != f {
			t.Errorf("%s: f.Pos(0) was not found in f", f.Name())
		}

		// add lines with SetLinesForContent and verify all positions
		src := test.source
		if src == nil {
			// no test source available - create one from scratch
			src = makeTestSource(test.size, test.lines)
		}
		f.SetLinesForContent(src)
		if f.LineCount() != len(test.lines) {
			t.Errorf("%s, SetLinesForContent: expected line count %d; got %d", f.Name(), len(test.lines), f.LineCount())
		}
		verifyPositions(t, fset, f, test.lines)

		verifyLineOffsets(t, fset, f, test.lines)
	}
}

func TestFiles(t *testing.T) {
	fset := NewFileSet()
	for i, test := range tests {
		base := fset.Base()
		if i%2 == 1 {
			// Setting a negative base is equivalent to
			// fset.Base(), so test some of each.
			base = -1
		}
		fset.AddFile(test.filename, base, test.size)
		j := 0
		fset.Iterate(func(f *File) bool {
			if f.Name() != tests[j].filename {
				t.Errorf("expected filename = %s; got %s", tests[j].filename, f.Name())
			}
			j++
			return true
		})
		if j != i+1 {
			t.Errorf("expected %d files; got %d", i+1, j)
		}
	}
}

// FileSet.File should return nil if Pos is past the end of the FileSet.
func TestFileSetPastEnd(t *testing.T) {
	fset := NewFileSet()
	for _, test := range tests {
		fset.AddFile(test.filename, fset.Base(), test.size)
	}
	if f := fset.File(Pos(fset.Base())); f != nil {
		t.Errorf("expected nil, got %v", f)
	}
}

func TestFileSetCacheUnlikely(t *testing.T) {
	fset := NewFileSet()
	offsets := make(map[string]int)
	for _, test := range tests {
		offsets[test.filename] = fset.Base()
		fset.AddFile(test.filename, fset.Base(), test.size)
	}
	for file, pos := range offsets {
		f := fset.File(Pos(pos))
		if f.Name() != file {
			t.Errorf("expecting %q at position %d, got %q", file, pos, f.Name())
		}
	}
}

// issue 4345. Test concurrent use of FileSet.Pos does not trigger a
// race in the FileSet position cache.
func TestFileSetRace(t *testing.T) {
	fset := NewFileSet()
	for i := 0; i < 100; i++ {
		fset.AddFile(fmt.Sprintf("file-%d", i), fset.Base(), 1031)
	}
	max := int32(fset.Base())
	var stop sync.WaitGroup
	r := rand.New(rand.NewSource(7))
	for i := 0; i < 2; i++ {
		r := rand.New(rand.NewSource(r.Int63()))
		stop.Add(1)
		go func() {
			for i := 0; i < 1000; i++ {
				fset.Position(Pos(r.Int31n(max)))
			}
			stop.Done()
		}()
	}
	stop.Wait()
}

func TestFileByteOffsetOfRune(t *testing.T) {
	fset := NewFileSet()
	b := []byte("x→y")
	f := fset.AddFile("f", fset.Base(), len(b))
	f.SetByteOffsetsForContent(b)

	xB := f.ByteOffsetOfRune(0)
	if want := 0; xB != want {
		t.Errorf("got `x` byte offset at %d, want %d", xB, want)
	}

	yB := f.ByteOffsetOfRune(2)
	if want := 4; yB != want {
		t.Errorf("got `y` byte offset at %d, want %d", yB, want)
	}
}
