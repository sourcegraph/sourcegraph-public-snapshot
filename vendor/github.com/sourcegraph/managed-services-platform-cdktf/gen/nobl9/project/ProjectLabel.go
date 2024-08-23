package project


type ProjectLabel struct {
	// A key for the label, unique within the associated resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/project#key Project#key}
	Key *string `field:"required" json:"key" yaml:"key"`
	// A list of unique values for a single key.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/project#values Project#values}
	Values *[]*string `field:"required" json:"values" yaml:"values"`
}

