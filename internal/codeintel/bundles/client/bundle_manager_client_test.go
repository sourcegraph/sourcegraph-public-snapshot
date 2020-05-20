package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSendUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("unexpected method. want=%s have=%s", "POST", r.Method)
		}
		if r.URL.Path != "/uploads/42" {
			t.Errorf("unexpected method. want=%s have=%s", "/uploads/42", r.URL.Path)
		}

		if content, err := ioutil.ReadAll(r.Body); err != nil {
			t.Fatalf("unexpected error reading payload: %s", err)
		} else if diff := cmp.Diff([]byte("payload\n"), content); diff != "" {
			t.Errorf("unexpected request payload (-want +got):\n%s", diff)
		}
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.SendUpload(context.Background(), 42, bytes.NewReader([]byte("payload\n")))
	if err != nil {
		t.Fatalf("unexpected error sending upload: %s", err)
	}
}

func TestSendUploadBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.SendUpload(context.Background(), 42, bytes.NewReader([]byte("payload\n")))
	if err == nil {
		t.Fatalf("unexpected nil error sending upload")
	}
}

func TestSendUploadPart(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("unexpected method. want=%s have=%s", "POST", r.Method)
		}
		if r.URL.Path != "/uploads/42/3" {
			t.Errorf("unexpected method. want=%s have=%s", "/uploads/42/3", r.URL.Path)
		}

		if content, err := ioutil.ReadAll(r.Body); err != nil {
			t.Fatalf("unexpected error reading payload: %s", err)
		} else if diff := cmp.Diff([]byte("payload\n"), content); diff != "" {
			t.Errorf("unexpected request payload (-want +got):\n%s", diff)
		}
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.SendUploadPart(context.Background(), 42, 3, bytes.NewReader([]byte("payload\n")))
	if err != nil {
		t.Fatalf("unexpected error sending upload: %s", err)
	}
}

func TestSendUploadPartBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.SendUploadPart(context.Background(), 42, 3, bytes.NewReader([]byte("payload\n")))
	if err == nil {
		t.Fatalf("unexpected nil error sending upload")
	}
}

func TestStitchParts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("unexpected method. want=%s have=%s", "POST", r.Method)
		}
		if r.URL.Path != "/uploads/42/stitch" {
			t.Errorf("unexpected method. want=%s have=%s", "/uploads/42/stitch", r.URL.Path)
		}
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.StitchParts(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error sending upload: %s", err)
	}
}

func TestStitchPartsBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.StitchParts(context.Background(), 42)
	if err == nil {
		t.Fatalf("unexpected nil error sending upload")
	}
}

func TestDeleteUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("unexpected method. want=%s have=%s", "POST", r.Method)
		}
		if r.URL.Path != "/uploads/42" {
			t.Errorf("unexpected method. want=%s have=%s", "/uploads/42", r.URL.Path)
		}
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.DeleteUpload(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error sending upload: %s", err)
	}
}

func TestDeleteUploadBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.DeleteUpload(context.Background(), 42)
	if err == nil {
		t.Fatalf("unexpected nil error sending upload")
	}
}

func TestGetUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "POST", r.Method)
		}
		if r.URL.Path != "/uploads/42" {
			t.Errorf("unexpected method. want=%s have=%s", "/uploads/42", r.URL.Path)
		}

		for i := 0; i < 1000; i++ {
			if _, err := w.Write([]byte(fmt.Sprintf("payload %d\n", i))); err != nil {
				t.Fatalf("unexpected error writing to client: %s", err)
			}
		}
	}))
	defer ts.Close()

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error creating temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	path, err := client.GetUpload(context.Background(), 42, tempDir)
	if err != nil {
		t.Fatalf("unexpected error sending db: %s", err)
	}

	if !strings.HasPrefix(path, tempDir) {
		t.Errorf("unexpected path location, want child of %s, got=%s", tempDir, path)
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading file: %s", err)
	}

	if lines := strings.Split(strings.TrimSpace(string(contents)), "\n"); len(lines) != 1000 {
		t.Errorf("unexpected payload size. want=%d have=%d", 1000, len(lines))
	}
}

func TestGetUploadNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	_, err := client.GetUpload(context.Background(), 42, "")
	if err != ErrNotFound {
		t.Fatalf("unexpected error. want=%q have=%q", ErrNotFound, err)
	}
}

func TestGetUploadBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	_, err := client.GetUpload(context.Background(), 42, "")
	if err == nil {
		t.Fatalf("unexpected nil error sending db")
	}
}

func TestSendDB(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("unexpected method. want=%s have=%s", "POST", r.Method)
		}
		if r.URL.Path != "/dbs/42" {
			t.Errorf("unexpected method. want=%s have=%s", "/dbs/42", r.URL.Path)
		}

		if content, err := ioutil.ReadAll(r.Body); err != nil {
			t.Fatalf("unexpected error reading payload: %s", err)
		} else if diff := cmp.Diff([]byte("payload\n"), content); diff != "" {
			t.Errorf("unexpected request payload (-want +got):\n%s", diff)
		}
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.SendDB(context.Background(), 42, bytes.NewReader([]byte("payload\n")))
	if err != nil {
		t.Fatalf("unexpected error sending db: %s", err)
	}
}

func TestSendDBBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	err := client.SendDB(context.Background(), 42, bytes.NewReader([]byte("payload\n")))
	if err == nil {
		t.Fatalf("unexpected nil error sending db")
	}
}

func TestBulkExists(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/exists" {
			t.Errorf("unexpected method. want=%s have=%s", "/exists", r.URL.Path)
		}

		if diff := cmp.Diff("1,2,3,4,5", r.URL.Query().Get("ids")); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}

		_, _ = w.Write([]byte(`{"1": false, "2": true, "3": false, "4": true, "5": true}`))
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	existsMap, err := client.Exists(context.Background(), []int{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatalf("unexpected error checking bulk exists: %s", err)
	}

	expected := map[int]bool{
		1: false,
		2: true,
		3: false,
		4: true,
		5: true,
	}
	if diff := cmp.Diff(expected, existsMap); diff != "" {
		t.Errorf("unexpected exists map (-want +got):\n%s", diff)
	}
}

func TestBulkExistsBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL}
	_, err := client.Exists(context.Background(), []int{1, 2, 3, 4, 5})
	if err == nil {
		t.Fatalf("unexpected nil error checking bulk exists")
	}
}
