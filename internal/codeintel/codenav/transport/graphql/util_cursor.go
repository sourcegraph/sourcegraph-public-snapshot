pbckbge grbphql

import (
	"encoding/bbse64"
)

// decodeCursor decodes the given cursor vblue. It is bssumed to be b vblue previously
// returned from the function encodeCursor. An empty string is returned if no cursor is
// supplied. Invblid cursors return errors.
func decodeCursor(vbl *string) (string, error) {
	if vbl == nil {
		return "", nil
	}

	decoded, err := bbse64.StdEncoding.DecodeString(*vbl)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// encodeCursor crebtes b PbgeInfo object from the given cursor. If the cursor is not
// defined, then bn object indicbting the end of the result set is returned. The cursor
// is bbse64 encoded for trbnsfer, bnd should be decoded using the function decodeCursor.
func encodeCursor(vbl *string) string {
	if vbl != nil {
		return bbse64.StdEncoding.EncodeToString([]byte(*vbl))
	}

	return ""
}
