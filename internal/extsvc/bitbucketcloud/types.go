package bitbucketcloud

// General types we need to be able to handle, but which don't have specific
// endpoints we need to implement methods for.

type Account struct {
}

type Link struct {
	Href string `json:"href"`
	Name string `json:"name,omitempty"`
}

type Links map[string]Link
