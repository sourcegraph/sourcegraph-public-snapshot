// Package PasswordPolciy implements a configurable password policy
package security

import (
	"errors"
	"regexp"
	"unicode/utf8"
	"strconv"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

// maxPasswordRunes is the maximum number of UTF-8 runes that a password can contain.
// This safety limit is to protect us from a DDOS attack caused by hashing very large passwords on Sourcegraph.com.
const maxPasswordRunes = 256

// Validate Password: Validates that a password meets the required criteria
func ValidatePassword(passwd string) error {

	if conf.ExperimentalFeatures().PasswordPolicy.Enabled {
		return validatePasswordUsingPolicy(passwd)
	}
	return validatePasswordUsingDefaultMethod(passwd)
}

// This is the default method using our current standard
func validatePasswordUsingDefaultMethod(passwd string) error {
	// Check for blank password
	if passwd == "" {
		return errors.New("password empty")
	}

	// Check for minimum/maximum length only
	pwLen := utf8.RuneCountInString(passwd)
	minPasswordRunes := conf.AuthMinPasswordLength()
	if pwLen < minPasswordRunes ||
		pwLen > maxPasswordRunes {
		return errcode.NewPresentationError(fmt.Sprintf("Password may not be less than %d or be more than %d characters.", minPasswordRunes, maxPasswordRunes))
		}

	return nil 
}

// This validates the password using the Paassord Policy configured
func validatePasswordUsingPolicy(passwd string) error {
	// Check for blank password
	if passwd == "" {
		return errors.New("password empty")
	}
	// Get a reference to the password policy 
	policy := conf.ExperimentalFeatures().PasswordPolicy
	
	// Minimum Length Check
	pwdLength := utf8.RuneCountInString(passwd)
	if pwdLength < policy.MinimumLength {
		return errcode.NewPresentationError(fmt.Sprintf("Password may not be less than %d characters.", policy.MinimumLength))
	}

	// Maximum Length Check
	if pwdLength > maxPasswordRunes {
		return errcode.NewPresentationError(fmt.Sprintf("Password may not be more than %d characters.", maxPasswordRunes))
	}

	// Numeric Check
	if policy.RequireAtLeastOneNumber {
		regex := regexp.MustCompile(`\d+`)
		numberFound := regex.MatchString(passwd)
		if !numberFound {
			return errors.New("Your password must include one number.")
		}
	}

	// Mixed case check
	if policy.RequireUpperandLowerCase {
		regexUpperCase := regexp.MustCompile(`[A-Z]+`)
		oneUpperCaseLetterFound := regexUpperCase.MatchString(passwd)
		if !oneUpperCaseLetterFound {
			return errors.New("Your password must include one uppercase letter.")
		}
		regexLowerCase := regexp.MustCompile(`[a-z]+`)
		oneLowerCaseLetterFound := regexLowerCase.MatchString(passwd)
		if !oneLowerCaseLetterFound {
			return errors.New("Your password must include one lowercase letter.")
		}
	}

	// Special Character Check
	if policy.NumberOfSpecialCharacters > 0 {
		regex := regexp.MustCompile(`\W` + `{` + strconv.Itoa(policy.NumberOfSpecialCharacters) + `}`)
		foundSpecialCharacters := regex.MatchString(passwd)
		if !foundSpecialCharacters {
			return errcode.NewPresentationError(fmt.Sprintf("Password must include at least %d special characters.", policy.NumberOfSpecialCharacters))
		}
	}

	// All good return
	return nil
}