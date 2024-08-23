package bigquerytable


type BigqueryTableExternalDataConfigurationJsonOptions struct {
	// The character encoding of the data.
	//
	// The supported values are UTF-8, UTF-16BE, UTF-16LE, UTF-32BE, and UTF-32LE. The default value is UTF-8.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#encoding BigqueryTable#encoding}
	Encoding *string `field:"optional" json:"encoding" yaml:"encoding"`
}

