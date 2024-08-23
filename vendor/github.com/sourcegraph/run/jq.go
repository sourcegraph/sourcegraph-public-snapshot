package run

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
)

// buildJQ parses and compiles a jq query.
func buildJQ(query string) (*gojq.Code, error) {
	jq, err := gojq.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("jq.Parse: %w", err)
	}
	jqCode, err := gojq.Compile(jq)
	if err != nil {
		return nil, fmt.Errorf("jq.Compile: %w", err)
	}
	return jqCode, nil
}

// execJQBytes can be used to execute a compiled jq query against small content bytes,
// e.g. lines. Errors are annotated with the provided content for ease of debugging.
func execJQBytes(ctx context.Context, jqCoode *gojq.Code, content []byte) ([]byte, error) {
	if len(content) == 0 {
		return nil, nil
	}
	result, err := execJQ(ctx, jqCoode, bytes.NewReader(content))
	if err != nil {
		// Embed the consumed content
		return nil, fmt.Errorf("%w: %s", err, string(content))
	}
	return result, nil
}

// execJQ executes the compiled jq query against content from reader.
func execJQ(ctx context.Context, jqCode *gojq.Code, reader io.Reader) ([]byte, error) {
	var input interface{}
	if err := json.NewDecoder(reader).Decode(&input); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}

	var result bytes.Buffer
	iter := jqCode.RunWithContext(ctx, input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("jq: %w", err)
		}

		encoded, err := gojq.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("jq: %w", err)
		}
		result.Write(encoded)
	}
	return result.Bytes(), nil
}
