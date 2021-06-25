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

func collectDocumentationPages(ctx context.Context, state *State) (chan *semantic.DocumentationPageData, chan *semantic.DocumentationPathInfoData) {
	pages := make(chan *semantic.DocumentationPageData, 1024)
	pathInfo := make(chan *semantic.DocumentationPathInfoData, 1024)
	if state.DocumentationResultRoot == -1 {
		close(pages)
		close(pathInfo)
		return pages, pathInfo
	}

	pageCollector := &pageCollector{
		numWorkers:                  32,
		isChildPage:                 false,
		state:                       state,
		parentPathID:                "",
		startingDocumentationResult: state.DocumentationResultRoot,
		dupChecker:                  &duplicateChecker{pathIDs: make(map[string]struct{}, 16*1024)},
		walkedPages:                 &duplicateChecker{pathIDs: make(map[string]struct{}, 128)},
	}

	tmpPages := make(chan *semantic.DocumentationPageData)
	go pageCollector.collect(ctx, tmpPages)

	go func() {
		// Emit path info for each page as a post-processing step once we've collected pages.
		for page := range tmpPages {
			var collectChildrenPages func(node *semantic.DocumentationNode) []string
			collectChildrenPages = func(node *semantic.DocumentationNode) []string {
				var children []string
				for _, child := range node.Children {
					if child.PathID != "" {
						children = append(children, child.PathID)
					} else if child.Node != nil {
						children = append(children, collectChildrenPages(child.Node)...)
					}
				}
				return children
			}
			isIndex := page.Tree.Label.Value == "" && page.Tree.Detail.Value == ""

			pages <- page
			pathInfo <- &semantic.DocumentationPathInfoData{
				PathID:   page.Tree.PathID,
				IsIndex:  isIndex,
				Children: collectChildrenPages(page.Tree),
			}
		}
		close(pages)
		close(pathInfo)
	}()
	return pages, pathInfo
}

type duplicateChecker struct {
	mu                        sync.RWMutex
	pathIDs                   map[string]struct{}
	duplicates, nonDuplicates int
}

func (d *duplicateChecker) add(pathID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.pathIDs[pathID]; ok {
		d.duplicates++
		return false
	}
	d.nonDuplicates++
	d.pathIDs[pathID] = struct{}{}
	return true
}

func (d *duplicateChecker) count() (duplicates, nonDupicates int) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.duplicates, d.nonDuplicates
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
	dupChecker, walkedPages     *duplicateChecker
}

func (p *pageCollector) collect(ctx context.Context, ch chan<- *semantic.DocumentationPageData) (remainingPages []*pageCollector) {
	var remainingPagesMu sync.RWMutex
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
			if !p.dupChecker.add(this.PathID) {
				log15.Warn("API docs: duplicate pathID forbidden", "pathID", this.PathID)
				return
			}
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
				if p.walkedPages.add(this.PathID) {
					remainingPagesMu.Lock()
					remainingPages = append(remainingPages, &pageCollector{
						isChildPage:                 true,
						parentPathID:                parent.PathID,
						state:                       p.state,
						startingDocumentationResult: documentationResult,
						dupChecker:                  p.dupChecker,
						walkedPages:                 p.walkedPages,
					})
					remainingPagesMu.Unlock()
				}
				return
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
			duplicates, nonDuplicates := p.dupChecker.count()
			if duplicates > 0 {
				log15.Error("API docs: upload failed due to duplicate pathIDs", "duplicates", duplicates, "nonDuplicates", nonDuplicates)
				return
			}
			ch <- &semantic.DocumentationPageData{Tree: this}
		}
	}
	walk(nil, p.startingDocumentationResult, p.parentPathID)
	if p.isChildPage {
		remainingPagesMu.RLock()
		defer remainingPagesMu.RUnlock()
		return remainingPages
	}

	// We are the root project page! Collect all the remaining pages.
	wg := &sync.WaitGroup{}
	remainingPagesMu.RLock()
	wg.Add(len(remainingPages))
	remainingPagesMu.RUnlock()
	for i := 0; i <= p.numWorkers; i++ {
		go func() {
			for {
				// Get a remaining page to process.
				remainingPagesMu.Lock()
				if len(remainingPages) == 0 { // HERE
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
