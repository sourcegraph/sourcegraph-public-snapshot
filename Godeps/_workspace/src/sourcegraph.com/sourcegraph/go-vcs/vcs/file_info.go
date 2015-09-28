package vcs

// ModeSubmodule is an os.FileMode mask indicating that the file is a
// VCS submodule (e.g., a git submodule).
const ModeSubmodule = 0160000

// SubmoduleInfo holds information about a VCS submodule and is
// returned in the FileInfo's Sys field by Stat/Lstat/ReadDir calls.
type SubmoduleInfo struct {
	// URL is the submodule repository origin URL.
	URL string

	// CommitID is the pinned commit ID of the submodule (in the
	// submodule repository's commit ID space).
	CommitID
}

// SymlinkInfo holds information about a symlink and is returned in
// the FileInfo's Sys field by Stat/Lstat/ReadDir calls.
type SymlinkInfo struct {
	// Dest is the path that the symlink points to.
	Dest string
}
