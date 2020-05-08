package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-multierror"
)

type ErrorPayload struct {
	Errors []GraphQLError `json:"errors"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

// graphQL performs GraphQL query on the frontend.
func graphQL(baseURL, token, query string, variables map[string]interface{}, target interface{}) error {
	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/.api/graphql", baseURL), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d\n%s\n", resp.StatusCode, contents)
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
