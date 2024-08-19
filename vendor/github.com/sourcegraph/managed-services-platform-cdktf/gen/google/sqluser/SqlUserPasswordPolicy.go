package sqluser


type SqlUserPasswordPolicy struct {
	// Number of failed attempts allowed before the user get locked.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#allowed_failed_attempts SqlUser#allowed_failed_attempts}
	AllowedFailedAttempts *float64 `field:"optional" json:"allowedFailedAttempts" yaml:"allowedFailedAttempts"`
	// If true, the check that will lock user after too many failed login attempts will be enabled.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#enable_failed_attempts_check SqlUser#enable_failed_attempts_check}
	EnableFailedAttemptsCheck interface{} `field:"optional" json:"enableFailedAttemptsCheck" yaml:"enableFailedAttemptsCheck"`
	// If true, the user must specify the current password before changing the password.
	//
	// This flag is supported only for MySQL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#enable_password_verification SqlUser#enable_password_verification}
	EnablePasswordVerification interface{} `field:"optional" json:"enablePasswordVerification" yaml:"enablePasswordVerification"`
	// Password expiration duration with one week grace period.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#password_expiration_duration SqlUser#password_expiration_duration}
	PasswordExpirationDuration *string `field:"optional" json:"passwordExpirationDuration" yaml:"passwordExpirationDuration"`
}

