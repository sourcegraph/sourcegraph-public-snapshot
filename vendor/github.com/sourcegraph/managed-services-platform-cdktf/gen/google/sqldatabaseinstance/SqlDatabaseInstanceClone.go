package sqldatabaseinstance


type SqlDatabaseInstanceClone struct {
	// The name of the instance from which the point in time should be restored.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#source_instance_name SqlDatabaseInstance#source_instance_name}
	SourceInstanceName *string `field:"required" json:"sourceInstanceName" yaml:"sourceInstanceName"`
	// The name of the allocated ip range for the private ip CloudSQL instance.
	//
	// For example: "google-managed-services-default". If set, the cloned instance ip will be created in the allocated range. The range name must comply with [RFC 1035](https://tools.ietf.org/html/rfc1035). Specifically, the name must be 1-63 characters long and match the regular expression [a-z]([-a-z0-9]*[a-z0-9])?.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#allocated_ip_range SqlDatabaseInstance#allocated_ip_range}
	AllocatedIpRange *string `field:"optional" json:"allocatedIpRange" yaml:"allocatedIpRange"`
	// (SQL Server only, use with point_in_time) clone only the specified databases from the source instance.
	//
	// Clone all databases if empty.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#database_names SqlDatabaseInstance#database_names}
	DatabaseNames *[]*string `field:"optional" json:"databaseNames" yaml:"databaseNames"`
	// The timestamp of the point in time that should be restored.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#point_in_time SqlDatabaseInstance#point_in_time}
	PointInTime *string `field:"optional" json:"pointInTime" yaml:"pointInTime"`
	// (Point-in-time recovery for PostgreSQL only) Clone to an instance in the specified zone.
	//
	// If no zone is specified, clone to the same zone as the source instance.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#preferred_zone SqlDatabaseInstance#preferred_zone}
	PreferredZone *string `field:"optional" json:"preferredZone" yaml:"preferredZone"`
}

