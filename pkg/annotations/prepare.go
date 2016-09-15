package annotations

import (
	"sort"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// Prepare should be called on annotations intended for the client to
// prepare them in ways described below for presentation in the UI.
func Prepare(anns []*sourcegraph.Annotation) []*sourcegraph.Annotation {
	if len(anns) == 0 {
		return anns
	}

	// Ensure that syntax highlighting is the innermost annotation so
	// that the CSS colors are applied (otherwise ref links appear in
	// the normal link color).
	for _, a := range anns {
		if a.URL == "" && len(a.URLs) == 0 {
			a.WantInner = 1
		}
	}

	sort.Sort(sortableAnnotations(anns))

	// Condense coincident refs ("multiple defs", such as an embedded Go
	// field's ref to both the field def and the type def).
	for i, ann := range anns {
		if i+1 == len(anns) {
			break
		}
		for j := i + 1; j < len(anns); j++ {
			ann2 := anns[j]
			if ann.StartByte == ann2.StartByte && ann.EndByte == ann2.EndByte {
				if (len(ann.URLs) > 0 || ann.URL != "") && (len(ann2.URLs) > 0 || ann2.URL != "") {
					if ann.URL != "" {
						ann.URLs = make([]string, 0, 2)
						ann.URLs = append(ann.URLs, ann.URL)
						ann.URL = ""
					}
					if ann2.URL != "" {
						ann.URLs = append(ann.URLs, ann2.URL)
					}
					ann.URLs = append(ann.URLs, ann2.URLs...)

					// Sort for determinism.
					sort.Strings(ann.URLs)

					// Delete the coincident ref.
					anns = append(anns[:j], anns[j+1:]...)
					j--
				}
			} else {
				break
			}
		}
	}

	return anns
}

type sortableAnnotations []*sourcegraph.Annotation

func (a sortableAnnotations) Len() int      { return len(a) }
func (a sortableAnnotations) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortableAnnotations) Less(i, j int) bool {
	return a[i].StartByte < a[j].StartByte ||
		(a[i].StartByte == a[j].StartByte && (a[i].EndByte < a[j].EndByte ||
			(a[i].EndByte == a[j].EndByte && (a[i].WantInner < a[j].WantInner))))
}
