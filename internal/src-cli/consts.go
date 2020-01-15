package srccli

// MinimumVersion is the minimum src-cli release version that works
// with that instance. This must be updated manually between releases.
// The public HTTP API will return that (or an updated patch version)
// as the suggested download with that instance.
//
// At the time of a Sourcegraph release, that is always the latest src-cli version.
const MinimumVersion = "3.9.0"
