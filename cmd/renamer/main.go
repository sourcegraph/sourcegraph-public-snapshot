package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"

	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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
	// Check if GQL API is 0-indexed.
	line      int
	character int
}

type codeRange struct {
	start codeLocation
	end   codeLocation
}

// loadSymbolLocations uses the GQL query from the scratchpad to query all locations for the given symbol.
func loadSymbolLocations(args programArgs) map[string][]codeRange {
	reqBody, err := json.Marshal(map[string]interface{}{"query": gqlSettingsQuery})
	if err != nil {
		return nil, errors.Wrap(err, "marshal request body")
	}

	url, err := gqlURL("CodeMonitorSettings")
	if err != nil {
		return nil, errors.Wrap(err, "construct frontend URL")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "construct request")
	}
	req.Header.Set("Content-Type", "application/json")
	if span != nil {
		carrier := opentracing.HTTPHeadersCarrier(req.Header)
		span.Tracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			carrier,
		)
	}

	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	var res gqlSettingsResponse
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
	return nil
}

// writeReplacement in-place replaces all the codeRanges in the given files by the replacement string.
func writeReplacement(ranges map[string][]codeRange, replacement string) error {

	// We need to make sure to order the codeRanges in ascending order and carry-forward
	// the offset of the replacement - original length to the next code ranges.
	// example line: func abc(a TYPE, b TYPE) error
	for filePath, crs := range ranges {
		f, err := os.OpenFile(filePath, os.O_RDWR, 0)
		if err != nil {
			return nil
		}
		io.ReadAll(context.Background(), f)
		for _, cr := range crs {

		}
	}
	return nil
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
