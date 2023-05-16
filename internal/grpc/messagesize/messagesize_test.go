package messagesize

import (
	"errors"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetMessageSizeBytesFromEnv(t *testing.T) {

	t.Run("8 MB", func(t *testing.T) {
		t.Setenv("TEST_SIZE", "8MB")

		size, err := getMessageSizeBytesFromEnv("TEST_SIZE", 0, math.MaxInt)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		expectedSize := 8 * 1000 * 1000
		if diff := cmp.Diff(expectedSize, size); diff != "" {
			t.Fatalf("unexpected size (-want +got):\n%s", diff)
		}
	})

	t.Run("invalid size", func(t *testing.T) {
		t.Setenv("TEST_SIZE", "this-is-not-a-size")

		_, err := getMessageSizeBytesFromEnv("TEST_SIZE", 0, math.MaxInt)
		var expectedErr *parseError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected parseError, got error %q", err)
		}
	})

	t.Run("just large enough", func(t *testing.T) {
		t.Setenv("TEST_SIZE", "4MB") // below range

		fourMegaBytes := 4 * 1000 * 1000
		size, err := getMessageSizeBytesFromEnv("TEST_SIZE", uint64(fourMegaBytes), math.MaxInt)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if diff := cmp.Diff(fourMegaBytes, size); diff != "" {
			t.Fatalf("unexpected size (-want +got):\n%s", diff)
		}
	})

	t.Run("just small enough", func(t *testing.T) {
		t.Setenv("TEST_SIZE", "4MB") // below range

		fourMegaBytes := 4 * 1000 * 1000
		size, err := getMessageSizeBytesFromEnv("TEST_SIZE", 0, uint64(fourMegaBytes))
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if diff := cmp.Diff(fourMegaBytes, size); diff != "" {
			t.Fatalf("unexpected size (-want +got):\n%s", diff)
		}
	})

	t.Run("too large", func(t *testing.T) {
		t.Setenv("TEST_SIZE", "4MB") // above range

		twoMegaBytes := 2 * 1024 * 1024
		_, err := getMessageSizeBytesFromEnv("TEST_SIZE", 0, uint64(twoMegaBytes))
		var expectedErr *sizeOutOfRangeError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected sizeOutOfRangeError, got error %q", err)
		}
	})

	t.Run("invalid size", func(t *testing.T) {
		t.Setenv("TEST_SIZE", "this-is-not-a-size")

		_, err := getMessageSizeBytesFromEnv("TEST_SIZE", 0, math.MaxInt)
		var expectedErr *parseError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected parseError, got error %q", err)
		}
	})

	t.Run("unset environment variable", func(t *testing.T) {
		// don't set the environment variable

		_, err := getMessageSizeBytesFromEnv("ALMOST_CERTAINLY_NOT_SET_ENV_VAR", 0, math.MaxInt)
		var expectedErr *envNotSetError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected envNotSetError, got error %q", err)
		}
	})
}
