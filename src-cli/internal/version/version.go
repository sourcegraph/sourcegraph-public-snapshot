package version

// DefaultBuildTag is the value BuildTag will be set to if this is not a release
// build.
const DefaultBuildTag = "dev"

// BuildTag is the git tag at the time of build and is used to
// denote the binary's current version. This value is supplied
// as an ldflag at compile time by the GoReleaser action.
var BuildTag = DefaultBuildTag
