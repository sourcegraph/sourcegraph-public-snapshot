package versions

// MockGetVersions, if non-nil, will be called instead of versions.GetVersions
var MockGetVersions func() ([]*Version, error)
