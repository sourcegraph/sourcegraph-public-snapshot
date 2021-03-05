package database

var (
	MockCannotCreateUserUsernameExistsErr = errCannotCreateUser{errorCodeUsernameExists}
	MockCannotCreateUserEmailExistsErr    = errCannotCreateUser{errorCodeEmailExists}
	MockUserNotFoundErr                   = userNotFoundErr{}
	MockUserEmailNotFoundErr              = userEmailNotFoundError{}
)
