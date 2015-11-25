package storageutil

import (
	"errors"
	"strings"
	"unicode"
)

// Note: these validation functions are tested at a higher level in store/testsuite/storage.go

// ValidateBucketName tells whether or not the bucket name is a valid one. It
// returns an error which should be presented to the user describing what is
// wrong with the name, or nil.
//
// An empty string is considered an error.
func ValidateBucketName(s string) error {
	if s == "" {
		return errors.New("bucket name may not be an empty string")
	}
	if strings.TrimSpace(s) != s {
		return errors.New("bucket name may not start or end with a space")
	}
	for _, r := range s {
		if r != '_' && r != '-' && r != '.' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return errors.New("bucket name may only contain underscores, dashes, periods, letters and digits")
		}
	}
	return nil
}

// ValidateAppName tells whether or not the app name is a valid one. It returns
// an error which should be presented to the user describing what is wrong with
// the name, or nil.
//
// An empty string is considered an error.
func ValidateAppName(s string) error {
	if s == "" {
		return errors.New("app name may not be an empty string")
	}
	if strings.TrimSpace(s) != s {
		return errors.New("app name may not start or end with a space")
	}
	for _, r := range s {
		if r != '_' && r != '-' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return errors.New("app name may only contain underscores, dashes, letters and digits")
		}
	}
	return nil
}
