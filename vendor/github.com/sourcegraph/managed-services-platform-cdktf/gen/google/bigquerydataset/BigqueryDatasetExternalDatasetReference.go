package bigquerydataset


type BigqueryDatasetExternalDatasetReference struct {
	// The connection id that is used to access the externalSource. Format: projects/{projectId}/locations/{locationId}/connections/{connectionId}.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#connection BigqueryDataset#connection}
	Connection *string `field:"required" json:"connection" yaml:"connection"`
	// External source that backs this dataset.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#external_source BigqueryDataset#external_source}
	ExternalSource *string `field:"required" json:"externalSource" yaml:"externalSource"`
}

