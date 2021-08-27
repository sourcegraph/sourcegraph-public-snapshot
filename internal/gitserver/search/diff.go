package search

import (
	"strings"
)

type Diff string

func (d Diff) ForEachDelta(f func(Delta) bool) {
	remaining := d
	var loc Location
	for len(remaining) > 0 {
		delta := scanDelta(string(remaining))
		remaining = remaining[len(delta):]

		newlineIdx := strings.IndexByte(delta, '\n')
		fileNameLine := delta[:newlineIdx]
		hunks := delta[newlineIdx+1:]
		fileNames := strings.Split(fileNameLine, " ")
		oldFile, newFile := fileNames[0], fileNames[1]

		if cont := f(Delta{
			location: loc,
			oldFile:  oldFile,
			newFile:  newFile,
			hunks:    hunks,
		}); !cont {
			return
		}

		loc = loc.Shift(Location{
			Offset: len(delta),
			Line:   strings.Count(delta, "\n"),
		})
	}
}

func scanDelta(s string) string {
	offset := 0
	for {
		idx := strings.IndexByte(s[offset:], '\n')
		if idx == -1 {
			return s
		}

		if idx+offset+1 == len(s) {
			return s
		}

		if strings.IndexByte("@+- <>=", s[idx+offset+1]) >= 0 {
			offset += idx + 1
		} else {
			return s[:offset+idx+1]
		}
	}
}

type Delta struct {
	location Location
	oldFile  string
	newFile  string
	hunks    string
}

func (d Delta) OldFile() (string, Location) {
	return d.oldFile, d.location
}

func (d Delta) NewFile() (string, Location) {
	return d.newFile, d.location.Shift(Location{
		Offset: len(d.newFile) + 1,
		Column: len(d.newFile) + 1,
	})
}

func (d Delta) ForEachHunk(f func(Hunk) bool) {
	remaining := d.hunks
	loc := d.location.Shift(Location{Line: 1, Offset: len(d.oldFile) + len(d.newFile) + len(" \n")})
	for len(remaining) > 0 {
		hunk := scanHunk(remaining)
		remaining = remaining[len(hunk):]

		newlineIdx := strings.IndexByte(hunk, '\n')
		header := hunk[:newlineIdx]
		lines := hunk[newlineIdx+1:]

		if cont := f(Hunk{
			location: loc,
			header:   header,
			lines:    lines,
		}); !cont {
			return
		}

		loc = loc.Shift(Location{
			Offset: len(hunk),
			Line:   strings.Count(hunk, "\n"),
		})
	}
}

func scanHunk(s string) string {
	offset := 0
	for {
		idx := strings.IndexByte(s[offset:], '\n')
		if idx == -1 {
			return s
		}

		if idx+offset+1 == len(s) {
			return s
		}

		switch s[idx+offset+1] {
		case '@':
			return s[:offset+idx+1]
		}
		offset += idx + 1
	}
}

type Hunk struct {
	location Location
	header   string
	lines    string
}

func (h Hunk) Header() (string, Location) {
	return h.header, h.location
}

func (h Hunk) ForEachLine(f func(Line) bool) {
	remaining := h.lines
	loc := h.location.Shift(Location{Line: 1, Offset: len(h.header) + len("\n")})
	for len(remaining) > 0 {
		line := scanLine(remaining)
		remaining = remaining[len(line):]

		if cont := f(Line{
			location: loc,
			fullLine: line,
		}); !cont {
			return
		}

		loc = loc.Shift(Location{
			Offset: len(line),
			Line:   1,
		})
	}
}

func scanLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx > 0 {
		return s[:idx+1]
	}
	return s
}

type Line struct {
	location Location
	fullLine string
}

func (l Line) Origin() byte {
	return l.fullLine[0]
}

func (l Line) Content() (string, Location) {
	return l.fullLine[1 : len(l.fullLine)-2], l.location.Shift(Location{Column: 1, Offset: 1})
}
