package lsifstore

import (
	"encoding/gob"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
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
	if err != nil {
		return nil, err
	}

	// encoding/gob does not retain the fact that an empty slice is not the same as a nil slice
	// (one encodes to `null` JSON) and we want to ensure we do not have `null` in our JSON lists.
	var walk func(*semantic.DocumentationNode)
	walk = func(node *semantic.DocumentationNode) {
		if node.Documentation.Tags == nil {
			node.Documentation.Tags = []protocol.Tag{}
		}
		if node.Children == nil {
			node.Children = []semantic.DocumentationNodeChild{}
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
func (s *Serializer) MarshalDocumentationPathInfoData(documentationPathInfo *semantic.DocumentationPathInfoData) ([]byte, error) {
	return s.encode(&documentationPathInfo)
}

// UnmarshalDocumentationPathInfoData is the inverse of MarshalDocumentationPathInfoData.
func (s *Serializer) UnmarshalDocumentationPathInfoData(data []byte) (documentationPathInfo *semantic.DocumentationPathInfoData, err error) {
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
