package app_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"src.sourcegraph.com/sourcegraph/pkg/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestRepoTree(t *testing.T) {
	c, mock := apptest.New()
	const source = "Milton"
	const expectedHTML = source

	mockRepoGet(mock, "my/repo")
	mockEmptyRepoConfig(mock)
	mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mockNoSrclibData(mock)
	calledRepoTreeGet := mockTreeEntryGet(mock, &sourcegraph.TreeEntry{
		SourceCode: &sourcegraph.SourceCode{
			Lines: []*sourcegraph.SourceCodeLine{
				{
					Tokens: []*sourcegraph.SourceCodeToken{
						&sourcegraph.SourceCodeToken{Class: "typ", Label: "Milton"},
					},
				},
			},
		},
		TreeEntry: &vcsclient.TreeEntry{
			Contents: []byte(source),
		},
	})

	resp, err := c.GetOK(router.Rel.URLToRepoTreeEntry("my/repo", "some/branch", "test.go").String())
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	contents := string(body)
	if !strings.Contains(contents, expectedHTML) {
		t.Errorf("Expected reponse body to contain '%s': %s", expectedHTML, contents)
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestRepoTree_markdown(t *testing.T) {
	c, mock := apptest.New()
	const docSource = "#Milton"
	const expectedHTML = "<h1>Milton</h1>"

	mockRepoGet(mock, "my/repo")
	mockEmptyRepoConfig(mock)
	mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mockNoSrclibData(mock)
	mockTreeEntryGet(mock, &sourcegraph.TreeEntry{
		TreeEntry: &vcsclient.TreeEntry{
			Contents: []byte(docSource),
		},
	})

	resp, err := c.GetOK(router.Rel.URLToRepoTreeEntry("my/repo", "some/branch", "test.md").String())
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	contents := string(body)
	if !strings.Contains(contents, expectedHTML) {
		t.Errorf("Expected reponse body to contain '%s'", expectedHTML)
	}
}

func TestRepoTree_plaintext(t *testing.T) {
	c, mock := apptest.New()
	const source = "Milton Woof"
	const expectedHTML = source

	mockRepoGet(mock, "my/repo")
	mockEmptyRepoConfig(mock)
	mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mockNoSrclibData(mock)
	mockTreeEntryGet(mock, &sourcegraph.TreeEntry{
		SourceCode: &sourcegraph.SourceCode{
			Lines: []*sourcegraph.SourceCodeLine{
				{
					Tokens: []*sourcegraph.SourceCodeToken{
						&sourcegraph.SourceCodeToken{Label: "Milton Woof"},
					},
				},
			},
		},
		TreeEntry: &vcsclient.TreeEntry{
			Contents: []byte(source),
		},
	})

	resp, err := c.GetOK(router.Rel.URLToRepoTreeEntry("my/repo", "some/branch", "filename.txt").String())
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	contents := string(body)
	if !strings.Contains(contents, expectedHTML) {
		t.Errorf("Expected reponse body to contain '%s': %s", expectedHTML, contents)
	}
}
