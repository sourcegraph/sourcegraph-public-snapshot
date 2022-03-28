package main

import (
	"errors"
	"flag"
	"os"
	"sort"
	"strings"

	"github.com/graphql-go/graphql/gqlerrors"
)

func main() {
	// Parse flags.

	// Craft GQL request.

	// Traverse response, write to files.
	println("hi rename is done")
}

type programArgs struct {
	repoName, revision, filePath string
	line, character              int
}

var errInvalidInput = errors.New("invalid input")

// parseFlags takes the input parameters of the program. The batch spec
// needs to pass all these values so we can uniquely identify the symbol.
// Errors if not all values are given.
func parseFlags() (args programArgs, replacement string, err error) {
	repoName := flag.String("repoName", "", "The repo name to change")
	if *repoName == "" {
		return programArgs{}, "", errInvalidInput
	}
	args.repoName = *repoName

	// todo: parse more args.

	return args, "", nil
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
func loadSymbolLocations(args programArgs) map[string][]codeRange {
	//reqBody, err := json.Marshal(map[string]interface{}{"query": gqlSettingsQuery})
	//if err != nil {
	//	return nil, errors.Wrap(err, "marshal request body")
	//}
	//
	//url, err := gqlURL("CodeMonitorSettings")
	//if err != nil {
	//	return nil, errors.Wrap(err, "construct frontend URL")
	//}
	//
	//req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	//if err != nil {
	//	return nil, errors.Wrap(err, "construct request")
	//}
	//req.Header.Set("Content-Type", "application/json")
	//if span != nil {
	//	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	//	span.Tracer().Inject(
	//		span.Context(),
	//		opentracing.HTTPHeaders,
	//		carrier,
	//	)
	//}
	//
	//resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	//if err != nil {
	//	return nil, errors.Wrap(err, "do request")
	//}
	//defer resp.Body.Close()
	//
	//var res gqlSettingsResponse
	//if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
	//	return nil, errors.Wrap(err, "decode response")
	//}
	//
	//if len(res.Errors) > 0 {
	//	var combined error
	//	for _, err := range res.Errors {
	//		combined = errors.Append(combined, err)
	//	}
	//	return nil, combined
	//}
	//return nil
	return nil
}

func writeReplacement(ranges map[string][]codeRange, replacement string) (err error) {
	for filePath, crs := range ranges {
		var buf []byte
		buf, err = os.ReadFile(filePath)
		content := string(buf)
		newCode, _ := replaceCode(content, crs, replacement) //TODO handle err
		println(newCode)
		// write that to file
	}
	return nil
}

// writeReplacement in-place replaces all the codeRanges in the given files by the replacement string.
func replaceCode(content string, ranges []codeRange, replacement string) (newCode string, err error) {

	// We need to make sure to order the codeRanges in ascending order and carry-forward
	// the offset of the replacement - original length to the next code ranges.
	// example line: func abc(a TYPE, b TYPE) error
	//TODO we think that end.line is always the same as start.line, we could ditch it

	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].start.line == ranges[j].start.line {
			return ranges[i].start.character < ranges[j].start.character
		}
		return ranges[i].start.line < ranges[j].start.line
	})

	lines := strings.Split(content, "\n")
	for _, cr := range ranges {
		line := lines[cr.start.line]
		line = line[:cr.start.character] + replacement + line[cr.end.character:]
		lines[cr.start.line] = line
	}

	return strings.Join(lines, "\n"), nil
}

const gqlSettingsQuery = `query CodeMonitorSettings{
	viewerSettings {
		final
	}
}`

type gqlSettingsResponse struct {
	Data struct {
		ViewerSettings struct {
			Final string `json:"final"`
		} `json:"viewerSettings"`
	} `json:"data"`
	Errors []gqlerrors.FormattedError
}
