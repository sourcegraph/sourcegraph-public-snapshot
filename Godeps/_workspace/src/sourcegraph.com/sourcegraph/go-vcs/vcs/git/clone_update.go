package git

// TODO: Fix the "gcrypt_init.c:5:1: warning: 'gcry_thread_cbs' is deprecated" warning
//       and remove -Wno-deprecated-declarations, if possible.

/*
#cgo CFLAGS: -Wno-deprecated-declarations
#cgo LDFLAGS: -lgcrypt
extern int _govcs_gcrypt_init();
*/
import "C"
import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"crypto/md5"

	"golang.org/x/crypto/ssh"

	git2go "github.com/libgit2/git2go"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
	sshutil "sourcegraph.com/sourcegraph/go-vcs/vcs/ssh"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/util"
)

func init() {
	// Overwrite the git cloner to use the faster libgit2
	// implementation.
	vcs.RegisterCloner("git", func(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error) {
		return Clone(url, dir, opt)
	})
}

func init() {
	// Initialize gcrypt for multithreaded operation. See
	// gcrypt_init.c for more information.
	rv := C._govcs_gcrypt_init()
	if rv != 0 {
		log.Fatal("gcrypt multithreaded init failed (see gcrypt_init.c)")
	}
}

func Clone(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error) {
	clopt := git2go.CloneOptions{Bare: opt.Bare}

	rc, cfs, err := makeRemoteCallbacks(url, opt.RemoteOpts)
	if err != nil {
		return nil, err
	}
	if cfs != nil {
		defer cfs.run()
	}
	if rc != nil {
		clopt.FetchOptions = &git2go.FetchOptions{RemoteCallbacks: *rc}
	}

	u, err := git2go.Clone(url, dir, &clopt)
	if err != nil {
		return nil, err
	}
	cr, err := gitcmd.Open(dir)
	if err != nil {
		return nil, err
	}
	r := &Repository{Repository: cr, u: u}
	if err := r.UpdateEverything(opt.RemoteOpts); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repository) UpdateEverything(opt vcs.RemoteOpts) error {
	// TODO(sqs): allow use of a remote other than "origin"
	rm, err := r.u.Remotes.Lookup("origin")
	if err != nil {
		return err
	}

	rc, cfs, err := makeRemoteCallbacks(rm.Url(), opt)
	if err != nil {
		return err
	}
	if cfs != nil {
		defer cfs.run()
	}
	var opts git2go.FetchOptions
	if rc != nil {
		opts.RemoteCallbacks = *rc
	}

	if err := rm.Fetch([]string{"+refs/*:refs/*"}, &opts, ""); err != nil {
		return err
	}

	return nil
}

type cleanupFuncs []func() error

func (f cleanupFuncs) run() error {
	for _, cf := range f {
		if err := cf(); err != nil {
			return err
		}
	}
	return nil
}

// makeRemoteCallbacks constructs the remote callbacks for libgit2
// remote operations. Currently the remote callbacks are trivial
// (empty) except when using an SSH remote.
//
// cleanupFuncs's run method should be called when the RemoteCallbacks
// struct is done being used. It is OK to ignore the error return.
func makeRemoteCallbacks(url string, opt vcs.RemoteOpts) (rc *git2go.RemoteCallbacks, cfs cleanupFuncs, err error) {
	defer func() {
		// Clean up if error; don't expect the caller to clean up if
		// we have a non-nil error.
		if err != nil {
			cfs.run()
		}
	}()

	if opt.SSH != nil {
		var privkeyFilename, pubkeyFilename string
		var privkeyFile, pubkeyFile *os.File
		var err error

		if opt.SSH.PrivateKey != nil {
			privkeyFilename, privkeyFile, err = util.WriteKeyTempFile(url, opt.SSH.PrivateKey)
			if err != nil {
				return nil, nil, err
			}
			cfs = append(cfs, privkeyFile.Close)
			cfs = append(cfs, func() error { return os.Remove(privkeyFile.Name()) })

			// Derive public key from private key if empty.
			if opt.SSH.PublicKey == nil {
				privKey, err := ssh.ParsePrivateKey(opt.SSH.PrivateKey)
				if err != nil {
					return nil, cfs, err
				}
				opt.SSH.PublicKey = ssh.MarshalAuthorizedKey(privKey.PublicKey())
			}

			pubkeyFilename, pubkeyFile, err = util.WriteKeyTempFile(url, opt.SSH.PublicKey)
			if err != nil {
				return nil, cfs, err
			}
			cfs = append(cfs, pubkeyFile.Close)
			cfs = append(cfs, func() error { return os.Remove(pubkeyFile.Name()) })
		}

		rc = &git2go.RemoteCallbacks{
			CredentialsCallback: git2go.CredentialsCallback(func(url string, usernameFromURL string, allowedTypes git2go.CredType) (git2go.ErrorCode, *git2go.Cred) {
				var username string
				if usernameFromURL != "" {
					username = usernameFromURL
				} else if opt.SSH.User != "" {
					username = opt.SSH.User
				} else {
					if username == "" {
						u, err := user.Current()
						if err == nil {
							username = u.Username
						}
					}
				}
				if allowedTypes&git2go.CredTypeSshKey != 0 && opt.SSH.PrivateKey != nil {
					rv, cred := git2go.NewCredSshKey(username, pubkeyFilename, privkeyFilename, "")
					return git2go.ErrorCode(rv), &cred
				}
				log.Printf("No authentication available for git URL %q.", url)
				rv, cred := git2go.NewCredDefault()
				return git2go.ErrorCode(rv), &cred
			}),
			CertificateCheckCallback: git2go.CertificateCheckCallback(func(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
				// libgit2 currently always returns valid=false. It
				// may return valid=true in the future if it checks
				// host keys using known_hosts, but let's ignore valid
				// so we don't get that behavior unexpectedly.

				if InsecureSkipCheckVerifySSH {
					return git2go.ErrOk
				}

				if cert == nil {
					return git2go.ErrNotFound
				}

				if cert.Hostkey.Kind&git2go.HostkeyMD5 > 0 {
					keys, found := standardKnownHosts.Lookup(hostname)
					if found {
						hostFingerprint := md5String(cert.Hostkey.HashMD5)
						for _, key := range keys {
							knownFingerprint := md5String(md5.Sum(key.Marshal()))
							if hostFingerprint == knownFingerprint {
								return git2go.ErrOk
							}
						}
					}
				}

				log.Printf("Invalid certificate for SSH host %s: %v.", hostname, cert)
				return git2go.ErrGeneric
			}),
		}
	}

	return rc, cfs, nil
}

// InsecureSkipCheckVerifySSH controls whether the client verifies the
// SSH server's certificate or host key. If InsecureSkipCheckVerifySSH
// is true, the program is susceptible to a man-in-the-middle
// attack. This should only be used for testing.
var InsecureSkipCheckVerifySSH bool

// standardKnownHosts contains known_hosts from the system known_hosts
// file and the user's known_hosts file.
var standardKnownHosts sshutil.KnownHosts

func init() {
	var err error
	standardKnownHosts, err = sshutil.ReadStandardKnownHostsFiles()
	if err != nil {
		log.Printf("Warning: failed to read standard SSH known_hosts files (%s). SSH host key checking will fail.", err)
	}
}

// md5String returns a formatted string representing the given md5Sum in hex
func md5String(md5Sum [16]byte) string {
	md5Str := fmt.Sprintf("% x", md5Sum)
	md5Str = strings.Replace(md5Str, " ", ":", -1)
	return md5Str
}
