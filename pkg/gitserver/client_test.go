package gitserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestDoListMulti(t *testing.T) {
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`["a"]`))
	}))
	u1, _ := url.Parse(s1.URL)

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`["b"]`))
	}))
	u2, _ := url.Parse(s2.URL)

	s3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	u3, _ := url.Parse(s3.URL)

	t.Run("ok", func(t *testing.T) {
		list, err := doListMulti(context.Background(), "", []string{u1.Host, u2.Host})
		if err != nil {
			t.Fatal(err)
		}
		if want := []string{"a", "b"}; !reflect.DeepEqual(list, want) {
			t.Errorf("got %q, want %q", list, want)
		}
	})

	t.Run("HTTP error and non-JSON body", func(t *testing.T) {
		if _, err := doListMulti(context.Background(), "", []string{u1.Host, u3.Host}); err == nil {
			t.Fatal("err == nil")
		}
	})
}
