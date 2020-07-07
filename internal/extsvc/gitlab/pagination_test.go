package gitlab

import (
	"context"
	"net/http"
	"testing"
)

func TestPaginatedResult(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	t.Run("HTTP error", func(t *testing.T) {
		c.httpClient = &mockHTTPEmptyResponse{statusCode: 500}

		pr := c.newPaginatedResult("DELETE", "/", func() interface{} { return struct{}{} })
		data, err := pr.next(ctx)
		if data != nil {
			t.Errorf("unexpected non-nil data: %+v", data)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("no X-Next-Page", func(t *testing.T) {
		// An unpaginated result should just be treated as a result with a
		// single page.
		c.httpClient = &mockHTTPResponseBody{responseBody: "42"}

		pr := c.newPaginatedResult("GET", "/", func() interface{} { return 0.0 })
		data, err := pr.next(ctx)
		if data == nil {
			t.Error("unexpected nil data")
		} else if want := 42.0; data.(float64) != want {
			t.Errorf("unexpected data value: want=%v; have=%v", want, data)
		}
		if err != nil {
			t.Errorf("unexpected non-nil err: %+v", err)
		}

		// Subsequent calls to next() should return an empty value and no error.
		data, err = pr.next(ctx)
		if data == nil {
			t.Error("unexpected nil data")
		} else if want := 0.0; data.(float64) != want {
			t.Errorf("unexpected data value: want=%v; have=%v", want, data)
		}
		if err != nil {
			t.Errorf("unexpected non-nil err: %+v", err)
		}
	})

	t.Run("normal operation", func(t *testing.T) {
		// An unpaginated result should just be treated as a result with a
		// single page.
		body := &mockHTTPResponseBody{
			header:       make(http.Header),
			responseBody: "42",
		}
		body.header.Add("X-Next-Page", "/two")
		c.httpClient = body

		// Get the first page.
		pr := c.newPaginatedResult("GET", "/", func() interface{} { return 0.0 })
		data, err := pr.next(ctx)
		if data == nil {
			t.Error("unexpected nil data")
		} else if want := 42.0; data.(float64) != want {
			t.Errorf("unexpected data value: want=%v; have=%v", want, data)
		}
		if err != nil {
			t.Errorf("unexpected non-nil err: %+v", err)
		}

		// Now get the second, which should also be the final page.
		body.responseBody = "43"
		body.header.Del("X-Next-Page")
		data, err = pr.next(ctx)
		if data == nil {
			t.Error("unexpected nil data")
		} else if want := 43.0; data.(float64) != want {
			t.Errorf("unexpected data value: want=%v; have=%v", want, data)
		}
		if err != nil {
			t.Errorf("unexpected non-nil err: %+v", err)
		}

		// Subsequent calls to next() should return an empty value and no error.
		data, err = pr.next(ctx)
		if data == nil {
			t.Error("unexpected nil data")
		} else if want := 0.0; data.(float64) != want {
			t.Errorf("unexpected data value: want=%v; have=%v", want, data)
		}
		if err != nil {
			t.Errorf("unexpected non-nil err: %+v", err)
		}

	})
}
