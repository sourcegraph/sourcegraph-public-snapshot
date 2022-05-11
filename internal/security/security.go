// Package security implements a configurable password policy
// This package may eventually get broken up as other packages are added.
package security

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// maxPasswordRunes is the maximum number of UTF-8 runes that a password can contain.
// This safety limit is to protect us from a DDOS attack caused by hashing very large passwords on Sourcegraph.com.
const maxPasswordRunes = 256

// Validate Password: Validates that a password meets the required criteria
func ValidatePassword(passwd string) error {

	if policy := conf.ExperimentalFeatures().PasswordPolicy; policy != nil && policy.Enabled {
		return validatePasswordUsingPolicy(passwd)
	}
	return validatePasswordUsingDefaultMethod(passwd)
}

// This is the default method using our current standard
func validatePasswordUsingDefaultMethod(passwd string) error {
	// Check for blank password
	if passwd == "" {
		return errcode.NewPresentationError("Your password may not be empty.")
	}

	// Check for minimum/maximum length only
	pwLen := utf8.RuneCountInString(passwd)
	minPasswordRunes := conf.AuthMinPasswordLength()
	if pwLen < minPasswordRunes ||
		pwLen > maxPasswordRunes {
		return errcode.NewPresentationError(fmt.Sprintf("Your password may not be less than %d or be more than %d characters.", minPasswordRunes, maxPasswordRunes))
	}

	return nil
}

// This validates the password using the Password Policy configured
func validatePasswordUsingPolicy(passwd string) error {
	letters := 0
	numbers := false
	upperCase := false
	special := 0

	for _, c := range passwd {
		switch {
		case unicode.IsNumber(c):
			numbers = true
			letters++
		case unicode.IsUpper(c):
			upperCase = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special++
			letters++
		case unicode.IsLetter(c) || c == ' ':
			letters++
		default:
			//ignore
		}
	}
	// Check for blank password
	if letters == 0 {
		return errors.New("password empty")
	}

	// Get a reference to the password policy
	policy := conf.ExperimentalFeatures().PasswordPolicy

	// Minimum Length Check
	if letters < policy.MinimumLength {
		return errcode.NewPresentationError(fmt.Sprintf("Your password may not be less than %d characters.", policy.MinimumLength))
	}

	// Maximum Length Check
	if letters > maxPasswordRunes {
		return errcode.NewPresentationError(fmt.Sprintf("Your password may not be more than %d characters.", maxPasswordRunes))
	}

	// Numeric Check
	if policy.RequireAtLeastOneNumber {
		if !numbers {
			return errcode.NewPresentationError("Your password must include one number.")
		}
	}

	// Mixed case check
	if policy.RequireUpperandLowerCase {
		if !upperCase {
			return errcode.NewPresentationError("Your password must include one uppercase letter.")
		}
	}

	// Special Character Check
	if policy.NumberOfSpecialCharacters > 0 {
		if special < policy.NumberOfSpecialCharacters {
			return errcode.NewPresentationError(fmt.Sprintf("Your password must include at least %d special character(s).", policy.NumberOfSpecialCharacters))
		}
	}

	// All good return
	return nil
}
