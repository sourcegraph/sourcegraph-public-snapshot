package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestImport(t *testing.T) {
	var attempt int
	var gotBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt != 2 {
			http.Error(w, "only attempt two works for this test", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to ready body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		gotBody = body
	}))
	defer ts.Close()

	zipFile, err := ioutil.TempFile("", "TestImport")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(zipFile.Name())
	zipFile.Write([]byte("hi"))
	zipFile.Close()

	err = importWithRetry(ts.URL, zipFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Once we get to attempt 2, it should work
	if attempt != 2 {
		t.Errorf("attempt = %d != 2", attempt)
	}

	if string(gotBody) != "hi" {
		t.Errorf("gotBody = %v != %v", string(gotBody), "hi")
	}
}
