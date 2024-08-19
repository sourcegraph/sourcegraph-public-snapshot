package sqldatabaseinstance


type SqlDatabaseInstanceSettingsDenyMaintenancePeriod struct {
	// End date before which maintenance will not take place.
	//
	// The date is in format yyyy-mm-dd i.e., 2020-11-01, or mm-dd, i.e., 11-01
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#end_date SqlDatabaseInstance#end_date}
	EndDate *string `field:"required" json:"endDate" yaml:"endDate"`
	// Start date after which maintenance will not take place.
	//
	// The date is in format yyyy-mm-dd i.e., 2020-11-01, or mm-dd, i.e., 11-01
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#start_date SqlDatabaseInstance#start_date}
	StartDate *string `field:"required" json:"startDate" yaml:"startDate"`
	// Time in UTC when the "deny maintenance period" starts on start_date and ends on end_date.
	//
	// The time is in format: HH:mm:SS, i.e., 00:00:00
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#time SqlDatabaseInstance#time}
	Time *string `field:"required" json:"time" yaml:"time"`
}

