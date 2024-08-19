package bigquerydataset


type BigqueryDatasetTimeouts struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#create BigqueryDataset#create}.
	Create *string `field:"optional" json:"create" yaml:"create"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#delete BigqueryDataset#delete}.
	Delete *string `field:"optional" json:"delete" yaml:"delete"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#update BigqueryDataset#update}.
	Update *string `field:"optional" json:"update" yaml:"update"`
}

