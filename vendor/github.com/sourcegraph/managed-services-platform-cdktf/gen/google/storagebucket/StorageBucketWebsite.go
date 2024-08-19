package storagebucket


type StorageBucketWebsite struct {
	// Behaves as the bucket's directory index where missing objects are treated as potential directories.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#main_page_suffix StorageBucket#main_page_suffix}
	MainPageSuffix *string `field:"optional" json:"mainPageSuffix" yaml:"mainPageSuffix"`
	// The custom object to return when a requested resource is not found.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#not_found_page StorageBucket#not_found_page}
	NotFoundPage *string `field:"optional" json:"notFoundPage" yaml:"notFoundPage"`
}

