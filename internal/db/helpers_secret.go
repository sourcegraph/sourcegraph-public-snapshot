package db

import (
	"reflect"

	intSecrets "github.com/sourcegraph/sourcegraph/internal/secrets"
)

func encryptColumns(columns []string, i interface{}) (interface{}, error) {

	sec := intSecrets.CryptObject

	e := reflect.ValueOf(&i).Elem()

	for i := 0; i < e.NumField(); i++ {
		field := e.Type().Field(i)
		if stringInSlice(field.Name, columns) {
			typeName := field.Type.Kind()
			value := reflect.ValueOf(&field)
			if typeName.String() == "string" {
				encString, err := sec.EncryptIfPossible(value.String())
				if err != nil {
					return nil, err
				}

				value.SetString(encString)

			} else if typeName.String() == "slice" { // []byte and json.RawMessage are identical
				encBytes, err := sec.EncryptBytesIfPossible(value.Bytes())
				if err != nil {
					return nil, err
				}
				value.SetBytes([]byte(encBytes))
			} else {
				panic("panic at the disco.")
			}
		}
	}

	return i, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
