package apidocs

import (
	"bytes"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func WalkDocumentationNode(node *precise.DocumentationNode, f func(node *precise.DocumentationNode) error) error {
	if err := f(node); err != nil {
		return err
	}

	for _, child := range node.Children {
		if child.Node != nil {
			if err := WalkDocumentationNode(child.Node, f); err != nil {
				return err
			}
		}
	}

	return nil
}

type Document struct {
	PathID  string
	Encoded string
}

func EncodeDocuments(pages []*precise.DocumentationPageData) []Document {
	var documents []Document
	var currentDocument bytes.Buffer
	for _, page := range pages {
		_ = WalkDocumentationNode(page.Tree, func(node *precise.DocumentationNode) error {
			if node.Documentation.SearchKey == "" {
				return nil
			}

			label := Truncate(node.Label.String(), 256) // 256 bytes, enough for ~100 characters in all languages

			var entry bytes.Buffer
			fmt.Fprintf(&entry, "%s\n", node.PathID)
			fmt.Fprintf(&entry, "%s\n", node.Documentation.SearchKey)
			fmt.Fprintf(&entry, "%s\n", label)
			fmt.Fprintf(&entry, "\n")
			if currentDocument.Len()+entry.Len() >= 1*1024*1024 {
				documents = append(documents, Document{PathID: page.Tree.PathID, Encoded: string(currentDocument.Bytes())})
				currentDocument.Reset()
			}
			currentDocument.Write(entry.Bytes())
			return nil
		})
		if currentDocument.Len() > 0 {
			documents = append(documents, Document{PathID: page.Tree.PathID, Encoded: string(currentDocument.Bytes())})
		}
	}
	return documents
}
