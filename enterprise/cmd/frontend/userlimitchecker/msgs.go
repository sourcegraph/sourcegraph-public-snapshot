package userlimitchecker

const (
	ACTIVE_LICENSE_ERR      = "could not get active license"
	ACTIVE_CHECKER_ERR      = "could not get checker with this license ID"
	USER_COUNT_ERR          = "could not get current user count on license"
	USER_LIMIT_ERR          = "could not get user limit on license"
	WITHIN_LIMIT_MSG        = "user count within limit; admin not alerted"
	EMAIL_RECENTLY_SENT_MSG = "email recently sent; admin not alerted"
	USERS_NOT_INCREASED_MSG = "user count has not increased; admin not alerted"
	ADMIN_EMAILS_ERR        = "could not get list of verified site admin emails"
	EMAIL_SEND_ERR          = "could not send email"
	LICENSES_ERR            = "could not get list of licenses"
	NO_LICENSE_ERR          = "no license is associated with this instance"
	USER_LIST_ERR           = "could not get list of users"
)
