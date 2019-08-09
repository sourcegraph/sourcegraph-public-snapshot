package extsvc

import (
	"encoding/json"
	"errors"
	"fmt"
)

func setJSONOrError(field **json.RawMessage, value interface{}) {
	if value == nil {
		*field = nil
		return
	}

	b, err := json.Marshal(value)
	if err != nil {
		b, _ = json.Marshal(struct {
			Error string `json:"__jsonError"`
		}{Error: err.Error()})
	}
	*field = (*json.RawMessage)(&b)
}

// SetAccountData sets the AccountData field to the (JSON-encoded) value. If an error occurs during
// JSON encoding, a JSON object describing the error is written to the field, instead.
func (d *ExternalAccountData) SetAccountData(v interface{}) {
	setJSONOrError(&d.AccountData, v)
}

// SetAuthData sets the AuthData field to the (JSON-encoded) value. If an error occurs during JSON
// encoding, a JSON object describing the error is written to the field, instead.
func (d *ExternalAccountData) SetAuthData(v interface{}) {
	setJSONOrError(&d.AuthData, v)
}

// GetAccountData reads the AccountData field into the value. The value should be a pointer type to
// the type that was passed to SetAccountData.
func (d *ExternalAccountData) GetAccountData(v interface{}) error {
	return getJSONOrError(d.AccountData, v)
}

// GetAuthData reads the AuthData field into the value. The value should be a pointer type to the
// type that was passed to SetAuthData.
func (d *ExternalAccountData) GetAuthData(v interface{}) error {
	return getJSONOrError(d.AuthData, v)
}

func getJSONOrError(field *json.RawMessage, v interface{}) error {
	if field == nil {
		return errors.New("field was nil")
	}

	if err := json.Unmarshal([]byte(*field), v); err != nil {
		var jsonErr jsonError
		if err := json.Unmarshal([]byte(*field), &jsonErr); err != nil {
			return fmt.Errorf("could not parse field as JSON: %s", err)
		}
		return errors.New(jsonErr.Error)
	}
	return nil
}

type jsonError struct {
	Error string `json:"__jsonError"`
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_792(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
