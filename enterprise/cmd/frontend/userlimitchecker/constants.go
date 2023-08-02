package userlimitchecker

import "time"

const (
	ACTIVE_LICENSE_ERR    = "could not get active license"
	ACTIVE_CHECKER_ERR    = "could not get checker with this license ID"
	USER_COUNT_ERR        = "could not get current user count on license"
	USER_LIMIT_ERR        = "could not get current user limit on license"
	WITHIN_LIMIT_MSG      = "user count within limit; admin not alerted"
	EMAIL_RECENTLY_SENT   = "email recently sent or user count unchanged; admin not alerted"
	ADMIN_EMAILS_ERR      = "could not get list of verified site admin emails"
	EMAIL_SEND_ERR        = "could not send email"
	LICENSES_ERR          = "could not get list of licenses"
	NO_LICENSE_ERR        = "no license is associated with this instance"
	USER_LIST_ERR         = "could not get list of users"
	LC_LOGGER_SCOPE       = "UserLimitChecker"
	LC_LOGGER_DESC        = "monitors user limit for active license"
	NINETY_PERCENT        = 90
	ONE_WEEK              = 7 * 24 * time.Hour
)
