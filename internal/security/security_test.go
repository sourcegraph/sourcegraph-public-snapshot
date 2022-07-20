package security

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestPasswordPolicy(t *testing.T) {

	setMockPasswordPolicyConfig(false, true, 12,
		10, 12, 2, true, true)

	t.Run("PasswordPolicy correctly detects deviating passwords", func(t *testing.T) {
		password := "sup3rstr0ngbutn0teno0ugh"
		assert.ErrorContains(t, ValidatePassword(password),
			"Your password must include one uppercase letter.")

		password = "id0hav3symb0lsn0w!!works?"
		assert.ErrorContains(t, ValidatePassword(password),
			"Your password must include one uppercase letter.")

		password = "Andn0w?!!"
		assert.ErrorContains(t, ValidatePassword(password),
			"Your password may not be less than 12 characters.")

		password = strings.Repeat("A", 259)
		assert.ErrorContains(t, ValidatePassword(password),
			"Your password may not be more than 256 characters.")
	})

	t.Run("PasswordPolicy detects correct passwords", func(t *testing.T) {
		setMockPasswordPolicyConfig(false, true, 12,
			10, 2, 2, true, true)

		password := "tH1smustCert@!inlybe0kthen?"
		assert.Nil(t, ValidatePassword(password))

	})
}
