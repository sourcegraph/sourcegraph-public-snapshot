package github

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
)

func TestUnmarshal(t *testing.T) {
	type result struct {
		FieldA string
		FieldB string
	}
	cases := map[string]string{
		// Valid
		`[]`:                                  "",
		`[{"FieldA": "hi"}]`:                  "",
		`[{"FieldA": "hi", "FieldB": "bye"}]`: "",

		// Error
		`[[]]`:            `graphql: cannot unmarshal at offset 2: before "[["; after "]]": json: cannot unmarshal array into Go value of type github.result`,
		`[{"FieldA": 1}]`: `graphql: cannot unmarshal at offset 13: before "[{\"FieldA\": 1"; after "}]": json: cannot unmarshal number`,
	}
	// Large body
	repeated := strings.Repeat(`{"FieldA": "hi", "FieldB": "bye"},`, 100)
	cases[fmt.Sprintf(`[%s {"FieldA": 1}, %s]`, repeated, repeated[:len(repeated)-1])] = `graphql: cannot unmarshal at offset 3414: before ", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"}, {\"FieldA\": 1"; after "}, {\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"b": json: cannot unmarshal number`

	for data, errStr := range cases {
		var a []result
		var b []result
		errA := json.Unmarshal([]byte(data), &a)
		errB := unmarshal([]byte(data), &b)

		if len(data) > 50 {
			data = data[:50] + "..."
		}

		if !reflect.DeepEqual(a, b) {
			t.Errorf("Expected the same result unmarshalling %v\na: %v\nb: %v", data, a, b)
		}
		if !reflect.DeepEqual(errA, errors.Cause(errB)) {
			t.Errorf("Expected the same underlying error unmarshalling %v\na: %v\nb: %v", data, errA, errB)
		}
		got := ""
		if errB != nil {
			got = errB.Error()
		}
		if !strings.HasPrefix(got, errStr) {
			t.Errorf("Unexpected error message %v\ngot:  %s\nwant: %s", data, got, errStr)
		}
	}
}

func TestNewRepoCache_GitHubDotCom(t *testing.T) {
	url, _ := url.Parse("https://www.github.com")
	token := "asdf"

	// github.com caches should:
	// (1) use githubProxyURL for the prefix hash rather than the given url
	// (2) have a TTL of 10 minutes
	key := sha256.Sum256([]byte(token + ":" + githubProxyURL.String()))
	prefix := "gh_repo:" + base64.URLEncoding.EncodeToString(key[:])
	got := NewRepoCache(url, token, "", 0)
	want := rcache.NewWithTTL(prefix, 600)
	if *got != *want {
		t.Errorf("TestNewRepoCache_GitHubDotCom: got %#v, want %#v", *got, *want)
	}
}

func TestNewRepoCache_GitHubEnterprise(t *testing.T) {
	url, _ := url.Parse("https://www.sourcegraph.com")
	token := "asdf"

	// GitHub Enterprise caches should:
	// (1) use the given URL for the prefix hash
	// (2) have a TTL of 30 seconds
	key := sha256.Sum256([]byte(token + ":" + url.String()))
	prefix := "gh_repo:" + base64.URLEncoding.EncodeToString(key[:])
	got := NewRepoCache(url, token, "", 0)
	want := rcache.NewWithTTL(prefix, 30)
	if *got != *want {
		t.Errorf("TestNewRepoCache_GitHubEnterprise: got %#v, want %#v", *got, *want)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_794(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
