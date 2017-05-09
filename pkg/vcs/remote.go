package vcs

// RemoteOpts configures interactions with a remote repository.
type RemoteOpts struct {
	SSH *SSHConfig `json:"ssh"` // ssh configuration for communication with the remote

	HTTPS *HTTPSConfig `json:"https"` // Optional HTTPS configuration for communication with the remote.
}

type SSHConfig struct {
	User       string `json:"user,omitempty"`      // ssh user (if empty, inferred from URL)
	PublicKey  []byte `json:"publicKey,omitempty"` // ssh public key (if nil, inferred from PrivateKey)
	PrivateKey []byte `json:"privateKey"`          // ssh private key, usually passed to ssh.ParsePrivateKey (passphrases currently unsupported)
}

// HTTPSConfig configures HTTPS for communication with remotes.
type HTTPSConfig struct {
	User string `json:"user"` // User is the username provided to the vcs.
	Pass string `json:"pass"` // Pass is the password provided to the vcs.
}
