package messagesize

import (
	"errors"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetMessageSizeBytesFromString(t *testing.T) {

	t.Run("8 MB", func(t *testing.T) {
		sizeString := "8MB"

		size, err := getMessageSizeBytesFromString(sizeString, 0, math.MaxInt)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		expectedSize := 8 * 1000 * 1000
		if diff := cmp.Diff(expectedSize, size); diff != "" {
			t.Fatalf("unexpected size (-want +got):\n%s", diff)
		}
	})

	t.Run("just small enough", func(t *testing.T) {
		sizeString := "4MB" // inside large-end of range

		fourMegaBytes := 4 * 1000 * 1000
		size, err := getMessageSizeBytesFromString(sizeString, 0, uint64(fourMegaBytes))
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if diff := cmp.Diff(fourMegaBytes, size); diff != "" {
			t.Fatalf("unexpected size (-want +got):\n%s", diff)
		}
	})

	t.Run("just large enough", func(t *testing.T) {
		sizeString := "4MB" // inside low-end of range

		fourMegaBytes := 4 * 1000 * 1000
		size, err := getMessageSizeBytesFromString(sizeString, uint64(fourMegaBytes), math.MaxInt)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if diff := cmp.Diff(fourMegaBytes, size); diff != "" {
			t.Fatalf("unexpected size (-want +got):\n%s", diff)
		}
	})

	t.Run("invalid size", func(t *testing.T) {
		sizeString := "this-is-not-a-size"

		_, err := getMessageSizeBytesFromString(sizeString, 0, math.MaxInt)
		var expectedErr *parseError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected parseError, got error %q", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		sizeString := ""

		_, err := getMessageSizeBytesFromString(sizeString, 0, math.MaxInt)
		var expectedErr *parseError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected parseError, got error %q", err)
		}
	})

	t.Run("too large", func(t *testing.T) {
		sizeString := "4MB" // above range

		twoMegaBytes := 2 * 1024 * 1024
		_, err := getMessageSizeBytesFromString(sizeString, 0, uint64(twoMegaBytes))
		var expectedErr *sizeOutOfRangeError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected sizeOutOfRangeError, got error %q", err)
		}
	})

	t.Run("too small", func(t *testing.T) {
		sizeString := "1MB" // below range

		twoMegaBytes := 2 * 1024 * 1024
		_, err := getMessageSizeBytesFromString(sizeString, uint64(twoMegaBytes), math.MaxInt)
		var expectedErr *sizeOutOfRangeError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected sizeOutOfRangeError, got error %q", err)
		}
	})
}
