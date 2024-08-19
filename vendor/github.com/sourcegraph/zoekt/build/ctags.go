// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package build

import (
	"bytes"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/ctags"
)

// Make sure all names are lowercase here, since they are normalized
var enryLanguageMappings = map[string]string{
	"c#": "c_sharp",
}

func normalizeLanguage(filetype string) string {
	normalized := strings.ToLower(filetype)
	if mapped, ok := enryLanguageMappings[normalized]; ok {
		normalized = mapped
	}

	return normalized
}

func parseSymbols(todo []*zoekt.Document, languageMap ctags.LanguageMap, parserBins ctags.ParserBinMap) error {
	monitor := newMonitor()
	defer monitor.Stop()

	var tagsToSections tagsToSections

	parser := ctags.NewCTagsParser(parserBins)
	defer parser.Close()

	for _, doc := range todo {
		if len(doc.Content) == 0 || doc.Symbols != nil {
			continue
		}

		zoekt.DetermineLanguageIfUnknown(doc)

		parserType := languageMap[normalizeLanguage(doc.Language)]
		if parserType == ctags.NoCTags {
			continue
		}

		// If the parser type is unknown, default to universal-ctags
		if parserType == ctags.UnknownCTags {
			parserType = ctags.UniversalCTags
		}

		monitor.BeginParsing(doc)
		es, err := parser.Parse(doc.Name, doc.Content, parserType)
		monitor.EndParsing(es)

		if err != nil {
			return err
		}
		if len(es) == 0 {
			continue
		}

		symOffsets, symMetaData, err := tagsToSections.Convert(doc.Content, es)
		if err != nil {
			return fmt.Errorf("%s: %v", doc.Name, err)
		}
		doc.Symbols = symOffsets
		doc.SymbolsMetaData = symMetaData
	}

	return nil
}

// overlaps finds the proper position to insert a zoekt.DocumentSection with
// "start and "end" into "symOffsets". It returns -1 if the new section overlaps
// with one of the existing ones.
func overlaps(symOffsets []zoekt.DocumentSection, start, end uint32) int {
	i := 0
	for i = len(symOffsets) - 1; i >= 0; i-- {
		// The most common case is that we exit here, because symOffsets is sorted by
		// construction and start is in many cases monotonically increasing.
		if start >= symOffsets[i].End {
			break
		}
		if end <= symOffsets[i].Start {
			continue
		}
		// overlap
		return -1
	}
	return i + 1
}

// tagsToSections contains buffers to be reused between conversions of bytes
// ranges to metadata. This is done to reduce pressure on the garbage
// collector.
type tagsToSections struct {
	nlsBuf []uint32
}

// Convert ctags entries to byte ranges (zoekt.DocumentSection) with
// corresponding metadata (zoekt.Symbol).
//
// This can not be called concurrently.
func (t *tagsToSections) Convert(content []byte, tags []*ctags.Entry) ([]zoekt.DocumentSection, []*zoekt.Symbol, error) {
	nls := t.newLinesIndices(content)
	symOffsets := make([]zoekt.DocumentSection, 0, len(tags))
	symMetaData := make([]*zoekt.Symbol, 0, len(tags))

	for _, t := range tags {
		if t.Line <= 0 {
			// Observed this with a .JS file.
			continue
		}
		lineIdx := t.Line - 1
		if lineIdx >= len(nls) {
			// Observed this with a .TS file.
			continue
		}

		lineOff := uint32(0)
		if lineIdx > 0 {
			lineOff = nls[lineIdx-1] + 1
		}

		end := nls[lineIdx]
		line := content[lineOff:end]

		// This is best-effort only. For short symbol names, we will often determine the
		// wrong offset.
		intraOff := bytes.Index(line, []byte(t.Name))
		if intraOff < 0 {
			// for Go code, this is very common, since
			// ctags barfs on multi-line declarations
			continue
		}

		start := lineOff + uint32(intraOff)
		endSym := start + uint32(len(t.Name))

		i := overlaps(symOffsets, start, endSym)
		if i == -1 {
			// Detected an overlap. Give up.
			continue
		}

		symOffsets = slices.Insert(symOffsets, i, zoekt.DocumentSection{
			Start: start,
			End:   endSym,
		})
		symMetaData = slices.Insert(symMetaData, i, &zoekt.Symbol{
			Sym:        t.Name,
			Kind:       t.Kind,
			Parent:     t.Parent,
			ParentKind: t.ParentKind,
		})
	}

	return symOffsets, symMetaData, nil
}

// newLinesIndices returns an array of all indexes of '\n' aswell as a final
// value for the length of the document.
func (t *tagsToSections) newLinesIndices(in []byte) []uint32 {
	// reuse nlsBuf between calls to tagsToSections.Convert
	out := t.nlsBuf
	if out == nil {
		out = make([]uint32, 0, len(in)/30)
	}

	finalEntry := uint32(len(in))
	off := uint32(0)
	for len(in) > 0 {
		i := bytes.IndexByte(in, '\n')
		if i < 0 {
			out = append(out, finalEntry)
			break
		}

		off += uint32(i)
		out = append(out, off)

		in = in[i+1:]
		off++
	}

	// save buffer for reuse
	t.nlsBuf = out[:0]

	return out
}

// monitorReportStuck is how long we need to be analysing a document before
// reporting it to stdout.
const monitorReportStuck = 10 * time.Second

// monitorReportStatus is how often we given status updates
const monitorReportStatus = time.Minute

type monitor struct {
	mu sync.Mutex

	lastUpdate time.Time

	start        time.Time
	totalSize    int
	totalSymbols int

	currentDocName       string
	currentDocSize       int
	currentDocStuckCount int

	done chan struct{}
}

func newMonitor() *monitor {
	start := time.Now()
	m := &monitor{
		start:      start,
		lastUpdate: start,
		done:       make(chan struct{}),
	}
	go m.run()
	return m
}

func (m *monitor) BeginParsing(doc *zoekt.Document) {
	now := time.Now()
	m.mu.Lock()
	m.lastUpdate = now

	// set current doc
	m.currentDocName = doc.Name
	m.currentDocSize = len(doc.Content)

	m.mu.Unlock()
}

func (m *monitor) EndParsing(entries []*ctags.Entry) {
	now := time.Now()
	m.mu.Lock()
	m.lastUpdate = now

	// update aggregate stats
	m.totalSize += m.currentDocSize
	m.totalSymbols += len(entries)

	// inform done if we warned about current document
	if m.currentDocStuckCount > 0 {
		log.Printf("symbol analysis for %s (size %d bytes) is done and found %d symbols", m.currentDocName, m.currentDocSize, len(entries))
		m.currentDocStuckCount = 0
	}

	// unset current document
	m.currentDocName = ""
	m.currentDocSize = 0

	m.mu.Unlock()
}

func (m *monitor) Stop() {
	close(m.done)
}

func (m *monitor) run() {
	stuckTicker := time.NewTicker(monitorReportStuck / 2) // half due to sampling theorem (nyquist)
	statusTicker := time.NewTicker(monitorReportStatus)

	defer stuckTicker.Stop()
	defer statusTicker.Stop()

	for {
		select {
		case <-m.done:
			now := time.Now()
			m.mu.Lock()
			log.Printf("symbol analysis finished for shard statistics: duration=%v symbols=%d bytes=%d", now.Sub(m.start).Truncate(time.Second), m.totalSymbols, m.totalSize)
			m.mu.Unlock()
			return

		case <-stuckTicker.C:
			now := time.Now()
			m.mu.Lock()
			running := now.Sub(m.lastUpdate).Truncate(time.Second)
			report := monitorReportStuck * (1 << m.currentDocStuckCount) // double the amount of time each time we report
			if m.currentDocName != "" && running >= report {
				m.currentDocStuckCount++
				log.Printf("WARN: symbol analysis for %s (%d bytes) has been running for %v", m.currentDocName, m.currentDocSize, running)
			}
			m.mu.Unlock()

		case <-statusTicker.C:
			now := time.Now()
			m.mu.Lock()
			log.Printf("DEBUG: symbol analysis still running for shard statistics: duration=%v symbols=%d bytes=%d", now.Sub(m.start).Truncate(time.Second), m.totalSymbols, m.totalSize)
			m.mu.Unlock()
		}
	}
}
