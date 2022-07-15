package database

var (
	MockCannotCreateUserUsernameExistsErr = errCannotCreateUser{errorCodeUsernameExists}
	MockCannotCreateUserEmailExistsErr    = errCannotCreateUser{errorCodeEmailExists}
	MockUserNotFoundErr                   = UserNotFoundErr{}
	MockUserEmailNotFoundErr              = userEmailNotFoundError{}
)
