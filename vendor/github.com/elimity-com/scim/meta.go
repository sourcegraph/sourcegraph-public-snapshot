package scim

// meta is a complex attribute containing resource metadata. All "meta" sub-attributes are assigned by the service
// provider (have a "mutability" of "readOnly"), and all of these sub-attributes have a "returned" characteristic of
// "default". This attribute SHALL be ignored when provided by clients.
type meta struct {
	// ResourceType is the name of the resource type of the resource.
	ResourceType string `json:"resourceType"`
	// Created is the "DateTime" that the resource was added to the service provider.
	Created string `json:"created,omitempty"`
	// LastModified is the most recent DateTime that the details of this resource were updated at the service provider.
	// If this resource has never been modified since its initial creation, the value MUST be the same as the value of
	// "created".
	LastModified string `json:"lastModified,omitempty"`
	// Location is the URI of the resource being returned. This value MUST be the same as the "Content-Location" HTTP
	// response header.
	Location string `json:"location"`
	// Version is the version of the resource being returned. This value must be the same as the entity-tag (ETag) HTTP
	// response header.
	Version string `json:"version,omitempty"`
}
