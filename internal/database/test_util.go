package database

func MockEmailExistsErr() error {
	return errCannotCreateUser{errorCodeEmailExists}
}

func MockUsernameExistsErr() error {
	return errCannotCreateUser{errorCodeEmailExists}
}

func strptr(s string) *string {
	return &s
}

func boolptr(b bool) *bool {
	return &b
}
