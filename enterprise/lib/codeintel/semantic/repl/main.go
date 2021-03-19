package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic/diff"
)

const helpMsg string = `
?                                                     help
count                                                 list dumps
paths <dump id>                                       list some paths from a dump
query <dump id> <path> <line> <column>                query dumps
index <path to repo root with dump.lsif>              index new dump
`

const helpMsgTODO string = `
patch <path to repo root with dump.lsif> <base ID>    patch dump
docRefs <dump id> <paths file>                        replace list of paths in paths file with list of docs referencing in dump
eq <dump id> <dump id>                                test two dumps for equality
cd <path>                                             change directory
`

func main() {
	var bundles []*semantic.GroupedBundleDataMaps

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("\n> ")
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "count":
			fmt.Println(len(bundles))
			break
		case "query":
			if len(fields) != 5 {
				fmt.Println("wrong number of args to q")
				break
			}
			dumpID, err := strconv.Atoi(fields[1])
			if err != nil {
				fmt.Println("first arg should be an integer")
				break
			}
			path := fields[2]
			line, err := strconv.Atoi(fields[3])
			if err != nil {
				fmt.Println("third arg should be an integer")
				break
			}
			column, err := strconv.Atoi(fields[4])
			if err != nil {
				fmt.Println("fourth arg should be an integer")
				break
			}

			err = queryBundle(bundles[dumpID], path, line, column)
			if err != nil {
				fmt.Println(helpMsg)
				break
			}
			break

		case "paths":
			if len(fields) != 2 {
				fmt.Println("expected 1 argument to index")
				break
			}
			bundleID, err := strconv.Atoi(fields[1])
			if err != nil {
				fmt.Println("first argument should be int")
				break
			}

			idx := 0
			for path := range bundles[bundleID].Documents {
				fmt.Println(path)
				idx++
				if idx > 4 {
					break
				}
			}

		case "index":
			if len(fields) != 2 && len(fields) != 3 {
				fmt.Println("expected 1 or 2 arguments to index")
				break
			}
			dumpPath := fields[1]
			projectRoot := filepath.Dir(dumpPath)
			if len(fields) == 3 {
				projectRoot = fields[2]
			}

			bundle, err := conversion.CorrelateLocalGit(context.Background(), dumpPath, projectRoot)
			if err != nil {
			}
			bundles = append(bundles, semantic.GroupedBundleDataChansToMaps(bundle))
			fmt.Printf("finished indexing dump %v\n", len(bundles)-1)
			break

		case "diff":
			if len(fields) != 3 {
				fmt.Printf("expected 2 arguments to diff, got %v\n", len(fields))
				break
			}
			gotID, err := strconv.Atoi(fields[1])
			if err != nil || gotID >= len(bundles) {
				fmt.Println("first argument should be bundle ID")
				break
			}
			wantID, err := strconv.Atoi(fields[2])
			if err != nil || wantID >= len(bundles) {
				fmt.Println("second argument should be bundle ID")
				break
			}

			fmt.Println(diff.Diff(bundles[gotID], bundles[wantID]))
			break

		case "patch":
			// TODO
			// if len(fields) != 3 {
			// 	fmt.Println("expected 2 arguments to patch")
			// 	break
			// }
			// root := fields[1]
			// baseID, err := strconv.Atoi(fields[2])
			// if err != nil {
			// 	fmt.Println("second argument should be int")
			// }

			// _, err = readBundle(root)
			// if err != nil {
			// 	fmt.Println(helpMsg)
			// 	break
			// }

			// break

		default:
			fmt.Println(helpMsg)
		}

		fmt.Printf("\n> ")
	}
}

func queryBundle(bundle *semantic.GroupedBundleDataMaps, path string, line, character int) error {
	results, err := semantic.Query(bundle, path, line, character)
	if err != nil {
		fmt.Printf("No data found at location")
		return err
	}
	for idx, result := range results {
		fmt.Printf("Result %d:\n", idx)
		for idx, definition := range result.Definitions {
			if idx >= 10 {
				fmt.Printf("Abridging definitions...\n")
				break
			}
			fmt.Printf("Definition:   %v:%v:(%v, %v)\n", definition.URI, definition.StartLine, definition.StartCharacter, definition.EndCharacter)
		}

		for idx, reference := range result.References {
			if idx >= 10 {
				fmt.Printf("Abridging references...\n")
				break
			}
			fmt.Printf("Reference:    %v:%v:(%v, %v)\n", reference.URI, reference.StartLine, reference.StartCharacter, reference.EndCharacter)
		}

		for idx, moniker := range result.Monikers {
			if idx >= 10 {
				fmt.Printf("Abridging monikers...\n")
			}
			fmt.Printf("Moniker:      %v:%v:%v\n", moniker.Scheme, moniker.Identifier, moniker.Kind)
		}

		fmt.Printf("Hover data:\n\n%v\n\n", result.Hover)

	}
	return nil
}

// finds all documents which have definition edges pointing into the argument list
func documentsReferencing(bundle *semantic.GroupedBundleDataMaps, paths []string) (_ []string, err error) {
	pathMap := map[string]struct{}{}
	for _, path := range paths {
		pathMap[path] = struct{}{}
	}

	resultIDs := map[semantic.ID]struct{}{}
	for _, resultChunk := range bundle.ResultChunks {
		for resultID, documentIDRangeIDs := range resultChunk.DocumentIDRangeIDs {
			for _, documentIDRangeID := range documentIDRangeIDs {
				// Skip results that do not point into one of the given documents
				if _, ok := pathMap[resultChunk.DocumentPaths[documentIDRangeID.DocumentID]]; !ok {
					continue
				}

				resultIDs[resultID] = struct{}{}
			}
		}
	}

	var pathsReferencing []string
	for path, document := range bundle.Documents {
		for _, r := range document.Ranges {
			if _, ok := resultIDs[r.DefinitionResultID]; ok {
				pathsReferencing = append(pathsReferencing, path)
				break
			}
		}
	}

	return pathsReferencing, nil
}
