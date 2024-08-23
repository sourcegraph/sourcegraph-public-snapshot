package secretmanagersecret


type SecretManagerSecretRotation struct {
	// Timestamp in UTC at which the Secret is scheduled to rotate.
	//
	// A timestamp in RFC3339 UTC "Zulu" format, with nanosecond resolution and up to nine fractional digits. Examples: "2014-10-02T15:01:23Z" and "2014-10-02T15:01:23.045123456Z".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#next_rotation_time SecretManagerSecret#next_rotation_time}
	NextRotationTime *string `field:"optional" json:"nextRotationTime" yaml:"nextRotationTime"`
	// The Duration between rotation notifications.
	//
	// Must be in seconds and at least 3600s (1h) and at most 3153600000s (100 years).
	// If rotationPeriod is set, 'next_rotation_time' must be set. 'next_rotation_time' will be advanced by this period when the service automatically sends rotation notifications.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#rotation_period SecretManagerSecret#rotation_period}
	RotationPeriod *string `field:"optional" json:"rotationPeriod" yaml:"rotationPeriod"`
}

