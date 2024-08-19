package datastreamconnectionprofile


type DatastreamConnectionProfileGcsProfile struct {
	// The Cloud Storage bucket name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#bucket DatastreamConnectionProfile#bucket}
	Bucket *string `field:"required" json:"bucket" yaml:"bucket"`
	// The root path inside the Cloud Storage bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#root_path DatastreamConnectionProfile#root_path}
	RootPath *string `field:"optional" json:"rootPath" yaml:"rootPath"`
}

