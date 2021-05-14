package conversion

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

// TODO(slimsag): future: today we do not consume state.DocumentationResultsByResultSet which will
// become important for e.g. letting one documentationResult link to another.

func collectDocumentationPages(ctx context.Context, state *State) chan *semantic.DocumentationPageData {
	ch := make(chan *semantic.DocumentationPageData)

	var (
		current *semantic.DocumentationNode
		walk    func(documentationResult int, isRoot bool, pathID string)
	)
	walk = func(documentationResult int, isRoot bool, pathID string) {
		labelID := state.DocumentationStringLabel[documentationResult]
		detailID := state.DocumentationStringDetail[documentationResult]
		documentation := state.DocumentationResultsData[documentationResult]
		this := &semantic.DocumentationNode{
			PathID:        pathID + "/" + documentation.Slug,
			Documentation: documentation,
			Label:         state.DocumentationStringsData[labelID],
			Detail:        state.DocumentationStringsData[detailID],
		}
		if isRoot || this.Documentation.NewPage {
			if current != nil {
				current.Children = append(current.Children, semantic.DocumentationNodeChild{
					PathID: this.PathID,
				})
				ch <- &semantic.DocumentationPageData{Tree: current}
			}
			current = this
		} else {
			current.Children = append(current.Children, semantic.DocumentationNodeChild{
				Node: this,
			})
		}

		children := state.DocumentationChildren[documentationResult]
		for _, child := range children {
			walk(child, false, this.PathID)
		}
		if isRoot {
			close(ch)
		}
	}
	if state.DocumentationResultRoot != -1 {
		go walk(state.DocumentationResultRoot, true, "")
	} else {
		close(ch)
	}
	return ch
}
