package emailaddrs

import (
	"errors"
	"reflect"
	"testing"
)

func TestObfuscateAndDeobfuscate(t *testing.T) {
	tests := []struct {
		input      string
		wantOutput string
	}{
		{"example.com", "yocNOSWWsWEEXyA="},
	}
	for _, test := range tests {
		output := obfuscate(test.input)
		if test.wantOutput != output {
			t.Errorf("%q: want output == %q, got %q", test.input, test.wantOutput, output)
		}
		input2, err := deobfuscate(output)
		if err != nil {
			t.Errorf("%q: deobfuscate error: %s", test.input, err)
			continue
		}
		if test.input != input2 {
			t.Errorf("%q: want deobfuscated output to be original input %q, got %q", test.input, test.input, input2)
		}
	}
}

func TestObfuscateAndDeobfuscateEmailAddrs(t *testing.T) {
	tests := []struct {
		email              string
		wantObfuscated     string
		wantObfuscateErr   error
		wantDeobfuscateErr error
	}{
		{"a@example.com", "a@-x-yocNOSWWsWEEXyA=", nil, nil},
		{"example.com", "", errors.New(`email has no '@': "example.com"`), nil},
		{"a@", "", errors.New(`email domain is empty: "a@"`), nil},
		{"@example.com", "", errors.New(`email user is empty: "@example.com"`), nil},
	}
	for _, test := range tests {
		obfuscated, err := Obfuscate(test.email)
		if !reflect.DeepEqual(err, test.wantObfuscateErr) {
			t.Errorf("%q: want obfuscate error %q, got %q", test.email, test.wantObfuscateErr, err)
			continue
		}
		if err != nil {
			continue
		}
		if test.wantObfuscated != obfuscated {
			t.Errorf("%q: want obfuscated == %q, got %q", test.email, test.wantObfuscated, obfuscated)
		}
		email2, err := Deobfuscate(obfuscated)
		if err != test.wantDeobfuscateErr {
			t.Errorf("%q: want deobfuscate error %q, got %q", test.email, test.wantDeobfuscateErr, err)
			continue
		}
		if err != nil {
			continue
		}
		if test.email != email2 {
			t.Errorf("%q: want deobfuscated obfuscated to be original email %q, got %q", test.email, test.email, email2)
		}
	}
}
