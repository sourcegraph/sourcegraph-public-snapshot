package constructs


// An entry in the construct metadata table.
type MetadataEntry struct {
	// The data.
	Data interface{} `field:"required" json:"data" yaml:"data"`
	// The metadata entry type.
	Type *string `field:"required" json:"type" yaml:"type"`
	// Stack trace at the point of adding the metadata.
	//
	// Only available if `addMetadata()` is called with `stackTrace: true`.
	// Default: - no trace information.
	//
	Trace *[]*string `field:"optional" json:"trace" yaml:"trace"`
}

