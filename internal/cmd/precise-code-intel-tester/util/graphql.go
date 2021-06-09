package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-multierror"
)

type ErrorPayload struct {
	Errors []GraphQLError `json:"errors"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

// QueryGraphQL performs GraphQL query on the frontend.
func QueryGraphQL(ctx context.Context, endpoint, token, query string, variables map[string]interface{}, target interface{}) error {
	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/.api/graphql", endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	// Note: We do not use req.Context(ctx) here as it causes the frontend
	// to output long error logs, which is very noisy under high concurrency.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var errorPayload ErrorPayload
	if err := json.Unmarshal(contents, &errorPayload); err == nil && len(errorPayload.Errors) > 0 {
		var combined error
		for _, err := range errorPayload.Errors {
			combined = multierror.Append(combined, fmt.Errorf("%s", err.Message))
		}

		return combined
	}

	return json.Unmarshal(contents, &target)
}
