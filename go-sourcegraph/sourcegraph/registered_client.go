package sourcegraph

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
)

// Spec returns c's RegisteredClientSpec.
func (c *RegisteredClient) Spec() RegisteredClientSpec {
	return RegisteredClientSpec{ID: c.ID}
}

// ClientNameOrDefault returns the client name, if non-empty, or else
// "Client XYZ", where "XYZ" is a prefix of the client ID.
func (c *RegisteredClient) ClientNameOrDefault() string {
	if c.ClientName != "" {
		return c.ClientName
	}

	abbrev := c.ID
	if max := 5; len(abbrev) > max {
		abbrev = abbrev[:5]
	}
	return fmt.Sprintf("Client %s", abbrev)
}

// MarshalText implements encoding.TextMarshaler.
func (c *RegisteredClientCredentials) MarshalText() ([]byte, error) {
	b64 := base64.StdEncoding
	return []byte(b64.EncodeToString([]byte(c.ID)) + ":" + b64.EncodeToString([]byte(c.Secret))), nil
}

// UnmarshalText implements encoding.TextMarshaler.
func (c *RegisteredClientCredentials) UnmarshalText(data []byte) error {
	i := bytes.Index(data, []byte(":"))
	if i == -1 {
		return errors.New("invalid encoding for RegisteredClientCredentials: no delimiter")
	}

	b64 := base64.StdEncoding
	id, err := b64.DecodeString(string(data[:i]))
	if err != nil {
		return err
	}
	secret, err := b64.DecodeString(string(data[i+1:]))
	if err != nil {
		return err
	}

	*c = RegisteredClientCredentials{ID: string(id), Secret: string(secret)}
	return nil
}
