package security

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordPolicy(t *testing.T) {

	authPolicyLength := 15
	authPolicySpChr := 2

	setMockPasswordPolicyConfig(false, true, authPolicyLength, authPolicySpChr,
		true, true)

	t.Run("PasswordPolicy correctly detects deviating passwords", func(t *testing.T) {
		password := "sup3rstr0ngbutn0teno0ugh"
		assert.ErrorContains(t, ValidatePassword(password),
			"Your password must include one uppercase letter.")

		password = "id0hav3symb0lsn0w!!works?"
		assert.ErrorContains(t, ValidatePassword(password),
			"Your password must include one uppercase letter.")

		password = "Andn0w?!!"
		err := fmt.Sprintf("Your password may not be less than %d characters.", authPolicyLength)
		assert.ErrorContains(t, ValidatePassword(password), err)

		password = strings.Repeat("A", 259)
		assert.ErrorContains(t, ValidatePassword(password),
			"Your password may not be more than 256 characters.")

		authPolicySpChr = 0
		setMockPasswordPolicyConfig(false, true, authPolicyLength, authPolicySpChr,
			false, false)
		password = "thisshouldnowpassaswell"
		assert.Nil(t, ValidatePassword(password))
	})

	t.Run("PasswordPolicy detects correct passwords", func(t *testing.T) {
		setMockPasswordPolicyConfig(false, true, 12,
			2, true, true)

		password := "tH1smustCert@!inlybe0kthen?"
		assert.Nil(t, ValidatePassword(password))

		password = strings.Repeat("A", 259)
		assert.ErrorContains(t, ValidatePassword(password),
			"Your password may not be more than 256 characters.")

		setMockPasswordPolicyConfig(false, true, 12,
			2, true, true)
		password = "tH1smustCert@!inlybe0kthen?"
		assert.Nil(t, ValidatePassword(password))
	})

}
