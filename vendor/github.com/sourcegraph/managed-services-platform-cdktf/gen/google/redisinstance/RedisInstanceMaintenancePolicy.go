package redisinstance


type RedisInstanceMaintenancePolicy struct {
	// Optional. Description of what this policy is for. Create/Update methods return INVALID_ARGUMENT if the length is greater than 512.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#description RedisInstance#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// weekly_maintenance_window block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#weekly_maintenance_window RedisInstance#weekly_maintenance_window}
	WeeklyMaintenanceWindow interface{} `field:"optional" json:"weeklyMaintenanceWindow" yaml:"weeklyMaintenanceWindow"`
}

