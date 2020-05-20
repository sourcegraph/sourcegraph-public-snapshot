package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func post(query string, path string) (GQLResult, error) {
	gqlRequest := GQLRequest{Query: gqlSearch, Variables: &GQLSearchVariable{SearchQuery: query}}
	b, err := json.Marshal(gqlRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %s", err)
	}
	client := &http.Client{}
	url := fmt.Sprintf("%s:%s", endpoint, path)
	request, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("response error: %s", err)
	}
	resultString, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(string(resultString), "{") {
		return "", errors.New(string(resultString))
	}
	var result interface{}
	err = json.Unmarshal(resultString, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}
