package gitutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

func ReadSymbolicRefFile(file string) (ref, commit string, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", "", err
	}
	data = bytes.TrimSpace(data)
	if bytes.HasPrefix(data, []byte("ref: ")) {
		return string(bytes.TrimPrefix(data, []byte("ref: "))), "", nil
	}
	if len(data) == 40 /* git commit SHA */ {
		return "", string(data), nil
	}
	return "", "", fmt.Errorf("invalid symbolic ref file %q: no 'ref: ' prefix (contents are: %q)", file, data)
}
