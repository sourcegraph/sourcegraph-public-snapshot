package vcsclient

import "sourcegraph.com/sourcegraph/go-vcs/vcs"

// CloneInfo is the information needed to clone a repository.
type CloneInfo struct {
	// VCS is the type of VCS (e.g., "git")
	VCS string

	// CloneURL is the remote URL from which to clone.
	CloneURL string

	// Additional options
	vcs.RemoteOpts
}

// FileWithRange is returned by GetFileWithOptions and includes the
// returned file's TreeEntry as well as the actual range of lines and
// bytes returned (based on the GetFileOptions parameters). That is,
// if Start/EndLine are set in GetFileOptions, this struct's
// Start/EndByte will be set to the actual start and end bytes of
// those specified lines, and so on for the other fields in
// GetFileOptions.
type FileWithRange struct {
	*TreeEntry
	FileRange // range of actual returned tree entry contents within file
}
