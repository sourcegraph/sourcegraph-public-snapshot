package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func serveRepoDocs(codeIntelResolver graphqlbackend.CodeIntelResolver) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		common, err := newCommon(w, r, "", serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request was handled
		}
		path, _ := mux.Vars(r)["Path"]

		if codeIntelResolver != nil {
			lsifTreeResolver, err := codeIntelResolver.GitBlobLSIFData(r.Context(), &graphqlbackend.GitBlobLSIFDataArgs{
				Repo:      common.Repo,
				Commit:    common.CommitID,
				Path:      "",
				ExactPath: false,
				ToolName:  "",
			})
			if err != nil {
				return errors.Wrap(err, "GitBlobLSIFData")
			}
			if lsifTreeResolver != nil {
				documentationPage, err := lsifTreeResolver.DocumentationPage(r.Context(), &graphqlbackend.LSIFDocumentationPageArgs{
					PathID: path,
				})
				if err == nil {
					treeJSON := []byte(documentationPage.Tree().Value.(string))
					var tree precise.DocumentationNode
					if err := json.Unmarshal(treeJSON, &tree); err != nil {
						return errors.Wrap(err, "Unmarshal")
					}
					target := &tree
					if r.URL.RawQuery != "" {
						target = findDocumentationNode(&tree, path+"#"+r.URL.RawQuery)
						if target == nil {
							// The section/symbol specified by the ?Query parameter does not exist.
							// This could happen for a few reasons:
							//
							// 1. It actually doesn't exist, e.g. it was removed in a commit to the repo.
							// 2. A URL parameter API docs does not understand has been injected. This is
							//    stupidly common, e.g.:
							//     a. `?toast=integrations` being injected after navigating to an API docs page
							//        and having to sign in first.
							//     b. `?_ga=2.892337.1256632002....` being injected by Google Analytics from the
							//        Sourcegraph blog or other sites.
							//     c. `?utm_source` being injected by marketing in various locations.
							//
							// Either way, we don't know what the URL query is at this point. We know it's not a
							// section/symbol of documentation we're aware of right now, so remove it from the URL.
							r.URL.RawQuery = ""
							http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
							return nil
						}
					}
					title := brandNameSubtitle(fmt.Sprintf("%s - %s API docs", target.Documentation.SearchKey, repoShortName(common.Repo.Name)))
					common.Title = title
					common.Metadata.ShowPreview = true
					common.Metadata.Title = title
					desc := markdownToDescriptionText(target.Detail.String())
					desc = strings.Replace(desc, "\n", " ", -1)
					desc = strings.Replace(desc, "\t", " ", -1)
					if len(desc) > 200 {
						desc = desc[:199] + "…"
					}
					if len(desc) > 0 {
						runes := []rune(desc)
						runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
						desc = string(runes)
					}
					common.Metadata.Description = fmt.Sprintf("%s API docs & usage examples; %s", target.Documentation.SearchKey, desc)
				}
			}
		}

		// TODO(apidocs): emit URL that points to another route capable of generating preview images for API docs.
		//common.Metadata.PreviewImage = "https://..."
		return renderTemplate(w, "app.html", common)
	}
}

func findDocumentationNode(node *precise.DocumentationNode, pathID string) *precise.DocumentationNode {
	if node.PathID == pathID {
		return node
	}
	for _, child := range node.Children {
		if child.Node != nil {
			if found := findDocumentationNode(child.Node, pathID); found != nil {
				return found
			}
		}
	}
	return nil
}

func markdownToDescriptionText(markdown string) string {
	var (
		v               []rune
		backticks       int
		insideCodeBlock bool
	)
	for _, r := range []rune(markdown) {
		if r == '`' {
			backticks++
			if backticks == 3 {
				insideCodeBlock = true
			} else if backticks == 6 {
				insideCodeBlock = false
			}
		} else if insideCodeBlock {
			// discard
		} else {
			backticks = 0
			v = append(v, r)
		}
	}
	result := string(v)
	result = strings.Replace(result, "\n", " ", -1)
	result = strings.Replace(result, "\t", " ", -1)
	before := len(result)
	for {
		result = strings.Replace(result, "  ", " ", -1)
		if len(result) == before {
			break
		}
		before = len(result)
	}
	return strings.TrimSpace(result)
}
