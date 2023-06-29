package graphql

import (
	"encoding/base64"
)

// decodeCursor decodes the given cursor value. It is assumed to be a value previously
// returned from the function encodeCursor. An empty string is returned if no cursor is
// supplied. Invalid cursors return errors.
func decodeCursor(val *string) (string, error) {
	if val == nil {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// encodeCursor creates a PageInfo object from the given cursor. If the cursor is not
// defined, then an object indicating the end of the result set is returned. The cursor
// is base64 encoded for transfer, and should be decoded using the function decodeCursor.
func encodeCursor(val *string) string {
	if val != nil {
		return base64.StdEncoding.EncodeToString([]byte(*val))
	}

	return ""
}
