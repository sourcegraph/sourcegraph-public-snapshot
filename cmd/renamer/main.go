package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/graphql-go/graphql/gqlerrors"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	// Parse flags.
	args, repoPath, replacement, err := parseFlags()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	// Craft GQL request.
	codeRanges, err := loadSymbolLocations(args)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	// Traverse response, write to files.
	err = writeReplacement(codeRanges, repoPath, replacement)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

type programArgs struct {
	repoName, revision, filePath string
	line, character              int
}

var errInvalidInput = errors.New("invalid input")

// parseFlags takes the input parameters of the program. The batch spec
// needs to pass all these values so we can uniquely identify the symbol.
// Errors if not all values are given.
func parseFlags() (args programArgs, repoPath string, replacement string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return programArgs{}, "", "", err
	}
	rp := flag.String("repoPath", "", cwd)
	repoName := flag.String("repoName", "", "")
	revision := flag.String("rev", "", "")
	filePath := flag.String("filePath", "", "")
	line := flag.Int("line", -1, "")
	character := flag.Int("character", -1, "")
	rep := flag.String("replacement", "", "")
	flag.Parse()

	if *rp == "" {
		return programArgs{}, "", "", errors.Wrap(errInvalidInput, "repoPath")
	}
	repoPath = *rp

	if *repoName == "" {
		return programArgs{}, "", "", errors.Wrap(errInvalidInput, "repoName")
	}
	args.repoName = *repoName

	if *revision == "" {
		return programArgs{}, "", "", errors.Wrap(errInvalidInput, "revision")
	}
	args.revision = *revision

	if *filePath == "" {
		return programArgs{}, "", "", errors.Wrap(errInvalidInput, "filePath")
	}
	args.filePath = *filePath

	if *line == -1 {
		return programArgs{}, "", "", errors.Wrap(errInvalidInput, "line")
	}
	args.line = *line

	if *character == -1 {
		return programArgs{}, "", "", errors.Wrap(errInvalidInput, "character")
	}
	args.character = *character

	if *rep == "" {
		return programArgs{}, "", "", errors.Wrap(errInvalidInput, "rep")
	}
	replacement = *rep

	return args, repoPath, replacement, nil
}

type codeLocation struct {
	// TODO Check if GQL API is 0-indexed (this code assumes YES).
	line      int
	character int
}

type codeRange struct {
	start codeLocation
	end   codeLocation
}

// loadSymbolLocations uses the GQL query from the scratchpad to query all locations for the given symbol.
func loadSymbolLocations(args programArgs) (map[string][]codeRange, error) {
	reqBody, err := json.Marshal(map[string]interface{}{"query": fmt.Sprintf(gqlQuery, args.repoName, args.revision, args.filePath, args.line, args.character)})
	if err != nil {
		return nil, errors.Wrap(err, "marshal request body")
	}

	url, err := gqlURL("LSIFReferencesForRename")
	if err != nil {
		return nil, errors.Wrap(err, "construct frontend URL")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "construct request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "src/renamer")
	req.Header.Set("Authorization", "token "+os.Getenv("SRC_ACCESS_TOKEN"))

	resp, err := http.DefaultClient.Do(req.WithContext(context.Background()))
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	var res gqlLSIFResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	if len(res.Errors) > 0 {
		var combined error
		for _, err := range res.Errors {
			combined = errors.Append(combined, err)
		}
		return nil, combined
	}

	if res.Data.Repository == nil {
		return nil, errors.New("repo not found")
	}
	if res.Data.Repository.Commit == nil {
		return nil, errors.New("commit not found")
	}
	if res.Data.Repository.Commit.Blob == nil {
		return nil, errors.New("blob not found")
	}
	if res.Data.Repository.Commit.Blob.LSIF == nil {
		return nil, errors.New("lsif data not found")
	}
	if res.Data.Repository.Commit.Blob.LSIF.References == nil {
		return nil, errors.New("lsif references not found")
	}

	crs := make(map[string][]codeRange)
	for _, ref := range res.Data.Repository.Commit.Blob.LSIF.References.Nodes {
		cr := codeRange{
			start: codeLocation{
				line:      ref.Range.Start.Line,
				character: ref.Range.Start.Character,
			},
			end: codeLocation{
				line:      ref.Range.End.Line,
				character: ref.Range.End.Character,
			},
		}
		if _, ok := crs[ref.Resource.Path]; !ok {
			crs[ref.Resource.Path] = make([]codeRange, 0)
		}
		crs[ref.Resource.Path] = append(crs[ref.Resource.Path], cr)
	}

	return crs, nil
}

// writeReplacement in-place replaces all the codeRanges in the given files by the replacement string.
func writeReplacement(ranges map[string][]codeRange, repoPath, replacement string) (err error) {
	for filePath, crs := range ranges {
		p := path.Join(repoPath, filePath)
		if !strings.HasPrefix(p, repoPath) {
			return errors.New("cannot change file outside of cwd")
		}
		var buf []byte
		buf, err = os.ReadFile(p)
		if err != nil {
			return err
		}
		content := string(buf)
		_, err := applyReplacement(content, crs, replacement)
		if err != nil {
			return err
		}
		fmt.Printf("I would write a replacement for %s\n", p)
		// if err := os.WriteFile(filePath, []byte(newCode), 0); err != nil {
		// 	return err
		// }
	}
	return nil
}

func applyReplacement(content string, ranges []codeRange, replacement string) (newCode string, err error) {

	// We need to make sure to order the codeRanges in ascending order and carry-forward
	// the offset of the replacement - original length to the next code ranges.
	// example line: func abc(a TYPE, b TYPE) error
	// TODO: we think that end.line is always the same as start.line, we could ditch it

	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].start.line == ranges[j].start.line {
			return ranges[i].start.character < ranges[j].start.character
		}
		return ranges[i].start.line < ranges[j].start.line
	})

	// peeking at the first element, as every other row should have an equal symbol length
	lengthDiff := len(replacement) - (ranges[0].end.character - ranges[0].start.character)

	lines := strings.Split(content, "\n")
	for idx, cr := range ranges {
		if cr.start.line != cr.end.line {
			return "", errors.New("unsupported multi-line rename")
		}
		if len(lines) < cr.start.line {
			return "", errors.Newf("tried to access line %d but only got %d", cr.start.line, len(lines))
		}
		line := lines[cr.start.line]

		// the first replacement can't be offset yet
		offset := 0
		if idx > 0 {
			offset = idx * lengthDiff
		}

		line = line[:cr.start.character+offset] + replacement + line[cr.end.character+offset:]
		lines[cr.start.line] = line
	}

	return strings.Join(lines, "\n"), nil
}

const gqlQuery = `query LSIFReferencesForRename {
	repository(name: %q) {
	  commit(rev: %q) {
		blob(path: %q) {
		  lsif {
			references(line: %d, character: %d) {
			  nodes {
				resource {
				  path
				}
				range {
				  start {
					line
					character
				  }
				  end {
					line
					character
				  }
				}
			  }
			}
		  }
		}
	  }
	}
  }
  `

type gqlLSIFResponse struct {
	Data struct {
		Repository *struct {
			Commit *struct {
				Blob *struct {
					LSIF *struct {
						References *struct {
							Nodes []*struct {
								Resource struct {
									Path string `json:"path"`
								} `json:"resource"`
								Range struct {
									Start struct {
										Line      int `json:"line"`
										Character int `json:"character"`
									} `json:"start"`
									End struct {
										Line      int `json:"line"`
										Character int `json:"character"`
									} `json:"end"`
								} `json:"range"`
							} `json:"nodes"`
						} `json:"references"`
					} `json:"lsif"`
				} `json:"blob"`
			} `json:"commit"`
		} `json:"repository"`
	} `json:"data"`
	Errors []gqlerrors.FormattedError
}

func gqlURL(queryName string) (string, error) {
	u, err := url.Parse("https://sourcegraph.test:3443")
	if err != nil {
		return "", err
	}
	u.Path = "/.api/graphql"
	u.RawQuery = queryName
	return u.String(), nil
}
