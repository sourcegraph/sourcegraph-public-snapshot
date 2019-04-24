package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/mattn/go-isatty"
)

// gqlURL returns the URL to the GraphQL endpoint for the given Sourcegraph
// instance.
func gqlURL(endpoint string) string {
	return endpoint + "/.api/graphql"
}

// apiRequest represents a GraphQL API request.
type apiRequest struct {
	query  string                 // the GraphQL query
	vars   map[string]interface{} // the GraphQL query variables
	result interface{}            // where to store the result
	done   func() error           // a function to invoke for handling the response. If nil, flags like -get-curl are ignored.
	endpoint string
	accessToken string

	// If true, errors will not be unpacked.
	//
	// Consider a GraphQL response like:
	//
	// 	{"data": {...}, "errors": ["something went really wrong"]}
	//
	// 'error unpacking' refers to how we will check if there are any `errors`
	// present in the response (if there are, we will report them on the command
	// line separately AND exit with a proper error code), and if there are no
	// errors `result` will contain only the `{...}` object.
	//
	// When true, the entire response object is stored in `result` -- as if you
	// ran the curl query yourself.
	dontUnpackErrors bool
}

// do performs the API request. Once the request is finished a.done is invoked to
// handle the response (which is stored in a.result).
func (a *apiRequest) do() error {
	// Create the JSON object.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(map[string]interface{}{
		"query":     a.query,
		"variables": a.vars,
	}); err != nil {
		return err
	}

	// Create the HTTP request.
	req, err := http.NewRequest("POST", gqlURL(a.endpoint), nil)
	if err != nil {
		return err
	}
	if a.accessToken != "" {
		req.Header.Set("Authorization", "token "+a.accessToken)
	}
	req.Body = ioutil.NopCloser(&buf)

	// Perform the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Our request may have failed before the reaching GraphQL endpoint, so
	// confirm the status code. You can test this easily with e.g. an invalid
	// endpoint like -endpoint=https://google.com
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized && isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			fmt.Println("You may need to specify or update your access token to use this endpoint.")
			fmt.Println("See https://github.com/sourcegraph/src-cli#authentication")
			fmt.Println("")
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error: %s\n\n%s", resp.Status, body)
	}

	// Decode the response.
	var result struct {
		Data   interface{} `json:"data,omitempty"`
		Errors interface{} `json:"errors,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Handle the case of not unpacking errors.
	if a.dontUnpackErrors {
		if err := jsonCopy(a.result, result); err != nil {
			return err
		}
		if err := a.done(); err != nil {
			return err
		}
		if result.Errors != nil {
			return &exitCodeError{error: nil, exitCode: graphqlErrorsExitCode}
		}
		return nil
	}

	// Handle the case of unpacking errors.
	if result.Errors != nil {
		return &exitCodeError{
			error:    fmt.Errorf("GraphQL errors:\n%s", &graphqlError{result.Errors}),
			exitCode: graphqlErrorsExitCode,
		}
	}
	if err := jsonCopy(a.result, result.Data); err != nil {
		return err
	}
	return a.done()
}

// jsonCopy is a cheaty method of copying an already-decoded JSON (src)
// response into its destination (dst) that would usually be passed to e.g.
// json.Unmarshal.
//
// We could do this with reflection, obviously, but it would be much more
// complex and JSON re-marshaling should be cheap enough anyway. Can improve in
// the future.
func jsonCopy(dst, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.NewDecoder(bytes.NewReader(data)).Decode(dst)
}

type graphqlError struct {
	Errors interface{}
}

func (g *graphqlError) Error() string {
	j, _ := marshalIndent(g.Errors)
	return string(j)
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
