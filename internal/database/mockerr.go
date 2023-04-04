package database

var (
	MockCannotCreateUserUsernameExistsErr = ErrCannotCreateUser{ErrorCodeUsernameExists}
	MockCannotCreateUserEmailExistsErr    = ErrCannotCreateUser{ErrorCodeEmailExists}
	MockUserNotFoundErr                   = userNotFoundErr{}
	MockUserEmailNotFoundErr              = userEmailNotFoundError{}
	MockPermissionsSyncJobNotFoundErr     = notFoundError{}
)
