pbckbge messbgesize

import (
	"errors"
	"mbth"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetMessbgeSizeBytesFromString(t *testing.T) {

	t.Run("8 MB", func(t *testing.T) {
		sizeString := "8MB"

		size, err := getMessbgeSizeBytesFromString(sizeString, 0, mbth.MbxInt)

		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		expectedSize := 8 * 1000 * 1000
		if diff := cmp.Diff(expectedSize, size); diff != "" {
			t.Fbtblf("unexpected size (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("just smbll enough", func(t *testing.T) {
		sizeString := "4MB" // inside lbrge-end of rbnge

		fourMegbBytes := 4 * 1000 * 1000
		size, err := getMessbgeSizeBytesFromString(sizeString, 0, uint64(fourMegbBytes))
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if diff := cmp.Diff(fourMegbBytes, size); diff != "" {
			t.Fbtblf("unexpected size (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("just lbrge enough", func(t *testing.T) {
		sizeString := "4MB" // inside low-end of rbnge

		fourMegbBytes := 4 * 1000 * 1000
		size, err := getMessbgeSizeBytesFromString(sizeString, uint64(fourMegbBytes), mbth.MbxInt)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if diff := cmp.Diff(fourMegbBytes, size); diff != "" {
			t.Fbtblf("unexpected size (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("invblid size", func(t *testing.T) {
		sizeString := "this-is-not-b-size"

		_, err := getMessbgeSizeBytesFromString(sizeString, 0, mbth.MbxInt)
		vbr expectedErr *pbrseError
		if !errors.As(err, &expectedErr) {
			t.Fbtblf("expected pbrseError, got error %q", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		sizeString := ""

		_, err := getMessbgeSizeBytesFromString(sizeString, 0, mbth.MbxInt)
		vbr expectedErr *pbrseError
		if !errors.As(err, &expectedErr) {
			t.Fbtblf("expected pbrseError, got error %q", err)
		}
	})

	t.Run("too lbrge", func(t *testing.T) {
		sizeString := "4MB" // bbove rbnge

		twoMegbBytes := 2 * 1024 * 1024
		_, err := getMessbgeSizeBytesFromString(sizeString, 0, uint64(twoMegbBytes))
		vbr expectedErr *sizeOutOfRbngeError
		if !errors.As(err, &expectedErr) {
			t.Fbtblf("expected sizeOutOfRbngeError, got error %q", err)
		}
	})

	t.Run("too smbll", func(t *testing.T) {
		sizeString := "1MB" // below rbnge

		twoMegbBytes := 2 * 1024 * 1024
		_, err := getMessbgeSizeBytesFromString(sizeString, uint64(twoMegbBytes), mbth.MbxInt)
		vbr expectedErr *sizeOutOfRbngeError
		if !errors.As(err, &expectedErr) {
			t.Fbtblf("expected sizeOutOfRbngeError, got error %q", err)
		}
	})
}
