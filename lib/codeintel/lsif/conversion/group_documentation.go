package conversion

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

// TODO(slimsag): future: today we do not consume state.DocumentationResultsByResultSet which will
// become important for e.g. letting one documentationResult link to another.

func collectDocumentationPages(ctx context.Context, state *State) chan *semantic.DocumentationPageData {
	ch := make(chan *semantic.DocumentationPageData)
	if state.DocumentationResultRoot == -1 {
		close(ch)
		return ch
	}

	pageCollector := &pageCollector{
		numWorkers:                  32,
		isChildPage:                 false,
		state:                       state,
		parentPathID:                "",
		startingDocumentationResult: state.DocumentationResultRoot,
	}
	if state.DocumentationResultRoot != -1 {
		pageCollector.startingDocumentationResult = state.DocumentationResultRoot
		go pageCollector.collect(ctx, ch)
	}
	return ch
}

// pageCollector collects all of the children for a single documentation page.
//
// It spawns a new pageCollector to collect each new child page as it encounters them.
type pageCollector struct {
	numWorkers                  int
	isChildPage                 bool
	parentPathID                string
	state                       *State
	startingDocumentationResult int
}

func (p *pageCollector) collect(ctx context.Context, ch chan<- *semantic.DocumentationPageData) (remainingPages []*pageCollector) {
	var walk func(parent *semantic.DocumentationNode, documentationResult int, pathID string)
	walk = func(parent *semantic.DocumentationNode, documentationResult int, pathID string) {
		isPageRoot := documentationResult == p.startingDocumentationResult

		labelID := p.state.DocumentationStringLabel[documentationResult]
		detailID := p.state.DocumentationStringDetail[documentationResult]
		documentation := p.state.DocumentationResultsData[documentationResult]
		this := &semantic.DocumentationNode{
			Documentation: documentation,
			Label:         p.state.DocumentationStringsData[labelID],
			Detail:        p.state.DocumentationStringsData[detailID],
		}
		if isPageRoot && !p.isChildPage {
			// We intentionally discard the project root PathID. This was the slug chosen by
			// the LSIF indexer, and has an arbitrary name like "/index". But we want a consistent
			// name "/" always for the root path ID: "/".
			this.PathID = "/"
		} else {
			this.PathID = pathID + documentation.Slug
		}

		if parent != nil && (isPageRoot || this.Documentation.NewPage) {
			// This documentationResult is a child of our parent, but it's a brand new page. We
			// spawn a new pageCollector to collect this page. We can't simply emit our page right
			// now, because we might not be finished collecting all the other descendant children
			// of this node.
			parent.Children = append(parent.Children, semantic.DocumentationNodeChild{
				PathID: this.PathID,
			})
			remainingPages = append(remainingPages, &pageCollector{
				isChildPage:                 true,
				parentPathID:                parent.PathID,
				state:                       p.state,
				startingDocumentationResult: documentationResult,
			})
		} else if parent != nil {
			parent.Children = append(parent.Children, semantic.DocumentationNodeChild{
				Node: this,
			})
		}

		children := p.state.DocumentationChildren[documentationResult]
		for _, child := range children {
			walk(this, child, this.PathID)
		}
		if isPageRoot {
			// collected a whole page
			ch <- &semantic.DocumentationPageData{Tree: this}
		}
	}
	walk(nil, p.startingDocumentationResult, p.parentPathID)
	if p.isChildPage {
		return remainingPages
	}

	// We are the root project page! Collect all the remaining pages.
	var (
		remainingPagesMu sync.RWMutex
		wg               = &sync.WaitGroup{}
	)
	wg.Add(len(remainingPages))
	for i := 0; i <= p.numWorkers; i++ {
		go func() {
			for {
				// Get a remaining page to process.
				remainingPagesMu.Lock()
				if len(remainingPages) == 0 {
					remainingPagesMu.Unlock()
					return // no more work
				}
				work := remainingPages[0]
				remainingPages = remainingPages[1:]
				remainingPagesMu.Unlock()

				// Perform work.
				newRemainingPages := work.collect(ctx, ch)

				// Add new work, if needed.
				if len(newRemainingPages) > 0 {
					wg.Add(len(newRemainingPages))
					remainingPagesMu.Lock()
					remainingPages = append(remainingPages, newRemainingPages...)
					remainingPagesMu.Unlock()
				}
				wg.Done()
			}
		}()
	}
	wg.Wait()

	close(ch) // collected all pages
	return nil
}
