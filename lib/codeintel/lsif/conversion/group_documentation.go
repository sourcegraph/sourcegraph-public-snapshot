package conversion

import (
	"context"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"

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
		dupChecker:                  &duplicateChecker{pathIDs: make(map[string]struct{}, 16*1024)},
	}
	if state.DocumentationResultRoot != -1 {
		pageCollector.startingDocumentationResult = state.DocumentationResultRoot
		go pageCollector.collect(ctx, ch)
	}
	return ch
}

type duplicateChecker struct {
	pathIDs                   map[string]struct{}
	duplicates, nonDuplicates int
}

func (d *duplicateChecker) check(pathID string) bool {
	if _, ok := d.pathIDs[pathID]; ok {
		d.duplicates++
		return true
	}
	d.nonDuplicates++
	d.pathIDs[pathID] = struct{}{}
	return false
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
	dupChecker                  *duplicateChecker
}

func (p *pageCollector) collect(ctx context.Context, ch chan<- *semantic.DocumentationPageData) (remainingPages []*pageCollector) {
	var walk func(parent *semantic.DocumentationNode, documentationResult int, pathID string)
	walk = func(parent *semantic.DocumentationNode, documentationResult int, pathID string) {
		labelID := p.state.DocumentationStringLabel[documentationResult]
		detailID := p.state.DocumentationStringDetail[documentationResult]
		documentation := p.state.DocumentationResultsData[documentationResult]
		this := &semantic.DocumentationNode{
			Documentation: documentation,
			Label:         p.state.DocumentationStringsData[labelID],
			Detail:        p.state.DocumentationStringsData[detailID],
		}
		switch {
		case pathID == "":
			this.PathID = "/"
		case this.Documentation.NewPage && pathID == "/":
			this.PathID = "/" + cleanPathIDElement(documentation.Identifier)
		case this.Documentation.NewPage:
			this.PathID = pathID + "/" + cleanPathIDElement(documentation.Identifier)
		default:
			this.PathID = pathID + "#" + cleanPathIDFragment(documentation.Identifier)
		}
		if p.dupChecker.check(this.PathID) {
			log15.Warn("API docs: duplicate pathID forbidden", "pathID", this.PathID)
			return
		}
		if parent != nil {
			if this.Documentation.NewPage {
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
					dupChecker:                  p.dupChecker,
				})
			} else {
				parent.Children = append(parent.Children, semantic.DocumentationNodeChild{
					Node: this,
				})
			}
		}

		children := p.state.DocumentationChildren[documentationResult]
		for _, child := range children {
			walk(this, child, pathIDTrimHash(this.PathID))
		}
		if documentationResult == p.startingDocumentationResult {
			// collected a whole page
			if p.dupChecker.duplicates > 0 {
				log15.Error("API docs: upload failed due to duplicate pathIDs", "duplicates", p.dupChecker.duplicates, "nonDuplicates", p.dupChecker.nonDuplicates)
				return
			}
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

// cleanPathIDElement replaces characters that may not be in URL path elements with dashes.
//
// It is not exhaustive, it only handles some common conflicts.
func cleanPathIDElement(s string) string {
	s = strings.Replace(s, "/", "-", -1)
	s = strings.Replace(s, "#", "-", -1)
	return s
}

// cleanPathIDFragment replaces characters that may not be in URL hashes with dashes.
//
// It is not exhaustive, it only handles some common conflicts.
func cleanPathIDFragment(s string) string {
	return strings.Replace(s, "#", "-", -1)
}

func joinPathIDs(a, b string) string {
	return a + "/" + b
}

func pathIDTrimHash(pathID string) string {
	i := strings.Index(pathID, "#")
	if i >= 0 {
		return pathID[:i]
	}
	return pathID
}
