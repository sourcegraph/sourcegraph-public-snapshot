package config

const (
	ConfigmapName = "sourcegraph-appliance"

	AnnotationKeyManaged             = "appliance.sourcegraph.com/managed"
	AnnotationConditions             = "appliance.sourcegraph.com/conditions"
	AnnotationKeyCurrentVersion      = "appliance.sourcegraph.com/currentVersion"
	AnnotationKeyConfigHash          = "appliance.sourcegraph.com/configHash"
	AnnotationKeyShouldTakeOwnership = "appliance.sourcegraph.com/adopted"
)
