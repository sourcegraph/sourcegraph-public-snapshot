package main

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

type DiffReporter struct {
	path  cmp.Path
	lines []string
}

func (r *DiffReporter) PushStep(ps cmp.PathStep) {
	// fmt.Println("descending into", ps.String())
	r.path = append(r.path, ps)
}

func (r *DiffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		// format based on type
		r.lines = append(r.lines, "-"+strings.Repeat("  ", len(r.path))+r.path.Last().String()+": "+fmt.Sprintf("%+v", vx))
		r.lines = append(r.lines, "+"+strings.Repeat("  ", len(r.path))+r.path.Last().String()+": "+fmt.Sprintf("%+v", vy))
	} else if !rs.ByIgnore() {
		vx, _ := r.path.Last().Values()
		r.lines = append(r.lines, strings.Repeat("  ", len(r.path))+r.path.Last().String()+": "+fmt.Sprintf("%+v", vx))
	}
}

func (r *DiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *DiffReporter) String() string {
	return strings.Join(r.lines, "\n")
}
