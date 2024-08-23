package computeinstance


type ComputeInstanceSchedulingLocalSsdRecoveryTimeout struct {
	// Span of time at a resolution of a second. Must be from 0 to 315,576,000,000 inclusive.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#seconds ComputeInstance#seconds}
	Seconds *float64 `field:"required" json:"seconds" yaml:"seconds"`
	// Span of time that's a fraction of a second at nanosecond resolution.
	//
	// Durations less than one second are represented
	// with a 0 seconds field and a positive nanos field. Must
	// be from 0 to 999,999,999 inclusive.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#nanos ComputeInstance#nanos}
	Nanos *float64 `field:"optional" json:"nanos" yaml:"nanos"`
}

