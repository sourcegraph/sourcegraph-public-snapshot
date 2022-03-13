package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type VersionRequester interface {
	LastCommit() (string, error)
	Environment() string
}

type APIVersionRequester struct {
	environment string
}

func NewAPIVersionRequester(environment string) *APIVersionRequester {
	return &APIVersionRequester{environment: environment}
}

func (a *APIVersionRequester) url() string {
	switch a.environment {
	case "preprod":
		return "https://preview.sgdev.dev/__version"
	case "production":
		return "https://sourcegraph.com/__version"
	default:
		return ""
	}
}

func (a *APIVersionRequester) Environment() string {
	return a.environment
}

func (a *APIVersionRequester) LastCommit() (string, error) {
	resp, err := http.Get(a.url())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	elems := strings.Split(string(body), "_")
	if len(elems) != 3 {
		return "", errors.Errorf("unknown format of /__version response: %q", body)
	}
	return elems[2], nil
}

type MockedVersionRequester struct {
	commit string
	err    error
}

func NewMockVersionRequester(lastCommit string, err error) *MockedVersionRequester {
	return &MockedVersionRequester{commit: lastCommit, err: err}
}

func (m *MockedVersionRequester) LastCommit() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.commit, nil
}

func (m *MockedVersionRequester) Environment() string {
	return "mock"
}
