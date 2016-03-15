package ui

import (
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

// StoreData contains preloaded data for the React Flux stores.
type StoreData struct {
	DefStore  DefStore
	BlobStore BlobStore
}

type DefStore struct {
	Defs map[string]*sourcegraph.Def `json:"defs"`
}

func (s *DefStore) AddDef(url string, def *sourcegraph.Def) {
	if s.Defs == nil {
		s.Defs = make(map[string]*sourcegraph.Def, 1)
	}
	s.Defs[url] = def
}

type BlobStore struct {
	Files       map[string]*sourcegraph.TreeEntry      `json:"files"`
	Annotations map[string]*sourcegraph.AnnotationList `json:"annotations"`
}

func (s *BlobStore) AddFile(spec sourcegraph.TreeEntrySpec, file *sourcegraph.TreeEntry) {
	if s.Files == nil {
		s.Files = make(map[string]*sourcegraph.TreeEntry, 1)
	}

	// This key logic must stay in sync with the keyFor func in
	// BlobStore.js.
	key := fmt.Sprintf("%s#%s#%s", spec.RepoRev.URI, spec.RepoRev.Rev, spec.Path)

	s.Files[key] = file
}

func (s *BlobStore) AddAnnotations(op *sourcegraph.AnnotationsListOptions, anns *sourcegraph.AnnotationList) {
	if s.Annotations == nil {
		s.Annotations = make(map[string]*sourcegraph.AnnotationList, 1)
	}

	// This key logic must stay in sync with the keyForAnns func in
	// BlobStore.js.
	key := fmt.Sprintf("%s#%s#%s#%s#%d#%d",
		op.Entry.RepoRev.URI,
		op.Entry.RepoRev.Rev,
		"", // This is empty in the JS code as well.
		op.Entry.Path,
		op.Range.StartByte,
		op.Range.EndByte,
	)

	s.Annotations[key] = anns
}
