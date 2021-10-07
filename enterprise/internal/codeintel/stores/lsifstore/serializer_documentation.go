package lsifstore

import (
	"encoding/gob"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func init() {
	gob.Register(&precise.DocumentationPageData{})
	gob.Register(&precise.DocumentationNode{})
	gob.Register(precise.DocumentationNodeChild{})
	gob.Register(&precise.DocumentationPathInfoData{})
}

// MarshalDocumentationPageData transforms documentation page data into a string of bytes writable to disk.
func (s *Serializer) MarshalDocumentationPageData(documentationPage *precise.DocumentationPageData) ([]byte, error) {
	return s.encode(&documentationPage)
}

// UnmarshalDocumentationPageData is the inverse of MarshalDocumentationPageData.
func (s *Serializer) UnmarshalDocumentationPageData(data []byte) (documentationPage *precise.DocumentationPageData, err error) {
	err = s.decode(data, &documentationPage)
	if err != nil {
		return nil, err
	}

	// encoding/gob does not retain the fact that an empty slice is not the same as a nil slice
	// (one encodes to `null` JSON) and we want to ensure we do not have `null` in our JSON lists.
	var walk func(*precise.DocumentationNode)
	walk = func(node *precise.DocumentationNode) {
		if node.Documentation.Tags == nil {
			node.Documentation.Tags = []protocol.Tag{}
		}
		if node.Children == nil {
			node.Children = []precise.DocumentationNodeChild{}
		}
		for _, child := range node.Children {
			if child.Node != nil {
				walk(child.Node)
			}
		}
	}
	walk(documentationPage.Tree)
	return documentationPage, nil
}

// MarshalDocumentationPathInfoData transforms documentation path info data into a string of bytes writable to disk.
func (s *Serializer) MarshalDocumentationPathInfoData(documentationPathInfo *precise.DocumentationPathInfoData) ([]byte, error) {
	return s.encode(&documentationPathInfo)
}

// UnmarshalDocumentationPathInfoData is the inverse of MarshalDocumentationPathInfoData.
func (s *Serializer) UnmarshalDocumentationPathInfoData(data []byte) (documentationPathInfo *precise.DocumentationPathInfoData, err error) {
	err = s.decode(data, &documentationPathInfo)
	if err != nil {
		return nil, err
	}

	// encoding/gob does not retain the fact that an empty slice is not the same as a nil slice
	// (one encodes to `null` JSON) and we want to ensure we do not have `null` in our JSON lists.
	if documentationPathInfo.Children == nil {
		documentationPathInfo.Children = []string{}
	}
	return documentationPathInfo, nil
}
