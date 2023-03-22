package graphql

import (
	"encoding/base64"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
)

// decodeCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeImplementationsCursor(rawEncoded string) (shared.ImplementationsCursor, error) {
	if rawEncoded == "" {
		return shared.ImplementationsCursor{Phase: "local"}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return shared.ImplementationsCursor{}, err
	}

	var cursor shared.ImplementationsCursor
	err = json.Unmarshal(raw, &cursor)
	return cursor, err
}

// encodeCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeImplementationsCursor(cursor shared.ImplementationsCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}

// decodeReferencesCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeReferencesCursor(rawEncoded string) (shared.ReferencesCursor, error) {
	if rawEncoded == "" {
		return shared.ReferencesCursor{Phase: "local"}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return shared.ReferencesCursor{}, err
	}

	var cursor shared.ReferencesCursor
	err = json.Unmarshal(raw, &cursor)
	return cursor, err
}

// encodeReferencesCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeReferencesCursor(cursor shared.ReferencesCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}

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
