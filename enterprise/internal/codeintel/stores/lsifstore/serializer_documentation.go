package lsifstore

import (
	"encoding/gob"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

func init() {
	gob.Register(&semantic.DocumentationNode{})
	gob.Register(semantic.DocumentationNodeChild{})
	gob.Register(&semantic.DocumentationPageData{})
}

// MarshalDocumentationPageData transforms documentation page data into a string of bytes writable to disk.
func (s *Serializer) MarshalDocumentationPageData(documentationPage semantic.DocumentationPageData) ([]byte, error) {
	return s.encode(&documentationPage)
}

// UnmarshalDocumentationPageData is the inverse of MarshalDocumentationPageData.
func (s *Serializer) UnmarshalDocumentationPageData(data []byte) (documentationPage semantic.DocumentationPageData, err error) {
	err = s.decode(data, &documentationPage)
	return documentationPage, err
}
