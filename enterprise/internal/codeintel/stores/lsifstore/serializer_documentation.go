package lsifstore

import (
	"encoding/gob"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

func init() {
	gob.Register(&semantic.DocumentationPageData{})
	gob.Register(&semantic.DocumentationNode{})
	gob.Register(semantic.DocumentationNodeChild{})
	gob.Register(&semantic.DocumentationPathInfoData{})
}

// MarshalDocumentationPageData transforms documentation page data into a string of bytes writable to disk.
func (s *Serializer) MarshalDocumentationPageData(documentationPage *semantic.DocumentationPageData) ([]byte, error) {
	return s.encode(&documentationPage)
}

// UnmarshalDocumentationPageData is the inverse of MarshalDocumentationPageData.
func (s *Serializer) UnmarshalDocumentationPageData(data []byte) (documentationPage *semantic.DocumentationPageData, err error) {
	err = s.decode(data, &documentationPage)
	return documentationPage, err
}

// MarshalDocumentationPathInfoData transforms documentation path info data into a string of bytes writable to disk.
func (s *Serializer) MarshalDocumentationPathInfoData(documentationPathInfo *semantic.DocumentationPathInfoData) ([]byte, error) {
	return s.encode(&documentationPathInfo)
}

// UnmarshalDocumentationPathInfoData is the inverse of MarshalDocumentationPathInfoData.
func (s *Serializer) UnmarshalDocumentationPathInfoData(data []byte) (documentationPathInfo *semantic.DocumentationPathInfoData, err error) {
	err = s.decode(data, &documentationPathInfo)
	return documentationPathInfo, err
}
