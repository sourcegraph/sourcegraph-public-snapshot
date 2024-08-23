package redisinstance


type RedisInstanceMaintenancePolicyWeeklyMaintenanceWindow struct {
	// Required. The day of week that maintenance updates occur.
	//
	// - DAY_OF_WEEK_UNSPECIFIED: The day of the week is unspecified.
	// - MONDAY: Monday
	// - TUESDAY: Tuesday
	// - WEDNESDAY: Wednesday
	// - THURSDAY: Thursday
	// - FRIDAY: Friday
	// - SATURDAY: Saturday
	// - SUNDAY: Sunday Possible values: ["DAY_OF_WEEK_UNSPECIFIED", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#day RedisInstance#day}
	Day *string `field:"required" json:"day" yaml:"day"`
	// start_time block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#start_time RedisInstance#start_time}
	StartTime *RedisInstanceMaintenancePolicyWeeklyMaintenanceWindowStartTime `field:"required" json:"startTime" yaml:"startTime"`
}

