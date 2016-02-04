package vcs

// RemoteOpts configures interactions with a remote repository.
type RemoteOpts struct {
	SSH *SSHConfig // ssh configuration for communication with the remote

	HTTPS *HTTPSConfig // Optional HTTPS configuration for communication with the remote.
}

type SSHConfig struct {
	User       string `json:",omitempty"` // ssh user (if empty, inferred from URL)
	PublicKey  []byte `json:",omitempty"` // ssh public key (if nil, inferred from PrivateKey)
	PrivateKey []byte // ssh private key, usually passed to ssh.ParsePrivateKey (passphrases currently unsupported)
}

// HTTPSConfig configures HTTPS for communication with remotes.
type HTTPSConfig struct {
	Pass string // Pass is the password provided to the vcs.
}

// A RemoteUpdater is a repository that can fetch updates to itself
// from a remote repository.
type RemoteUpdater interface {
	// UpdateEverything updates all branches, tags, etc., to match the
	// default remote repository. The implementation is VCS-dependent.
	//
	// If supported by the implementation, parsed results of update will be returned,
	// otherwise it'll be nil.
	UpdateEverything(RemoteOpts) (*UpdateResult, error)
}

// UpdateResult is the result of parsing output of the remote update operation.
type UpdateResult struct {
	Changes []Change
}

// Operation that happened to a branch.
type Operation uint8

const (
	// NewOp is a branch that was created.
	NewOp Operation = iota

	// FFUpdatedOp is a branch that was fast-forward updated.
	FFUpdatedOp

	// ForceUpdatedOp is a branch that was force updated.
	ForceUpdatedOp

	// DeletedOp is a branch that was deleted.
	DeletedOp
)

// Change is a single entry in the update result, representing Op done on Branch.
type Change struct {
	Op     Operation
	Branch string
}
