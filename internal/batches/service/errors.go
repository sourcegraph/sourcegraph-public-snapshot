package service

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ErrDuplicateCredential struct{}

func (e ErrDuplicateCredential) Error() string {
	return "a credential for this code host already exists"
}

func (e ErrDuplicateCredential) Extensions() map[string]any {
	return map[string]any{"code": "ErrDuplicateCredential"}
}

type ErrVerifyCredentialFailed struct {
	SourceErr error
}

func (e ErrVerifyCredentialFailed) Error() string {
	return fmt.Sprintf("Failed to verify the credential:\n```\n%s\n```\n", e.SourceErr)
}

func (e ErrVerifyCredentialFailed) Extensions() map[string]any {
	return map[string]any{"code": "ErrVerifyCredentialFailed"}
}

var VerifyCredentialTimeoutError = errors.New("verifying credential timed out")
