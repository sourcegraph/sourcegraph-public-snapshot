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
	User string // User is the username provided to the vcs.
	Pass string // Pass is the password provided to the vcs.
}
