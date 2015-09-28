/*
   Copyright 2014 CoreOS, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package ssh

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"crypto/hmac"
	"crypto/sha1"

	"golang.org/x/crypto/ssh"
)

const (
	sshHashDelim  = "|" // hostfile.h
	sshHashPrefix = "|1|"
)

// A KnownHost is a hostname and a known host key associated with that
// hostname. The hostname can be either unhashed or hashed.
type KnownHost struct {
	Hostnames []string // unhashed hostnames (represented as comma-separated names in the original file)

	Salt, Hash []byte // hashed hostname

	Key ssh.PublicKey
}

// Match returns whether hostname matches this known host entry's
// unhashed hostnames (separated by comma) or the hashed hostname.
func (h *KnownHost) Match(hostname string) bool {
	// TODO(sqs): lowercase before comparing? will that break hashed hostname lookups?

	for _, hn := range h.Hostnames {
		if hn == hostname {
			return true
		}
	}

	if h.Salt != nil && h.Hash != nil {
		mac := hmac.New(sha1.New, h.Salt)
		mac.Write([]byte(hostname))
		hash := mac.Sum(nil)
		if bytes.Equal(h.Hash, hash) {
			return true
		}
	}

	return false
}

// KnownHosts is a collection of known hosts and their host
// keys. Because hostname key may be hashed, use Lookup to get the
// host keys for a hostname instead of simply iterating over them and
// checking the Hostname field.
type KnownHosts []*KnownHost

// Lookup looks up hostname (which must be an unhashed hostname) in
// the known hosts collection. It returns host keys that match the
// unhashed hostname and the hashed variant of it. If any host keys
// are found, found is true; otherwise it is false.
func (khs KnownHosts) Lookup(hostname string) (hostKeys []ssh.PublicKey, found bool) {
	for _, h := range khs {
		if h.Match(hostname) {
			hostKeys = append(hostKeys, h.Key)
			found = true
		}
	}
	return hostKeys, found
}

// ReadStandardKnownHostsFiles reads and parses the known_hosts files
// at /etc/ssh/ssh_known_hosts and ~/.ssh/known_hosts.
func ReadStandardKnownHostsFiles() (KnownHosts, error) {
	// System known_hosts
	kh, err := ReadKnownHostsFile("/etc/ssh/known_hosts")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// User known_hosts
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	if u.HomeDir != "" {
		kh1, err := ReadKnownHostsFile(filepath.Join(u.HomeDir, ".ssh/known_hosts"))
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		kh = append(kh, kh1...)
	}

	return kh, nil
}

// ReadKnownHostsFile reads the known_hosts file at path.
func ReadKnownHostsFile(path string) (KnownHosts, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Check perms. TODO(sqs): we should really call Lstat on the path and not Stat here, and then avoid the TOCTTOU bug...
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Mode().Perm()&0022 > 0 {
		return nil, fmt.Errorf("known_hosts file %s must not be writable by others (mode %o)", path, fi.Mode().Perm())
	}

	return ParseKnownHosts(f)
}

// ParseKnownHosts parses an SSH known_hosts file.
func ParseKnownHosts(r io.Reader) (KnownHosts, error) {
	var khs KnownHosts
	s := bufio.NewScanner(r)
	n := 0
	for s.Scan() {
		n++
		line := s.Bytes()

		kh, err := parseKnownHostsLine(line)
		if err != nil {
			return nil, fmt.Errorf("parsing known_hosts: %s (line %d)", err, n)
		}
		if kh == nil {
			// empty line
			continue
		}

		khs = append(khs, kh)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return khs, nil
}

// parseKnownHostsLine parses a line from a known hosts file.  It
// returns a string containing the hosts section of the line, an
// ssh.PublicKey parsed from the line, and any error encountered
// during the parsing.
func parseKnownHostsLine(line []byte) (*KnownHost, error) {
	// Skip any leading whitespace.
	line = bytes.TrimLeft(line, "\t ")

	// Skip comments and empty lines.
	if bytes.HasPrefix(line, []byte("#")) || len(line) == 0 {
		return nil, nil
	}

	// Skip markers.
	if bytes.HasPrefix(line, []byte("@")) {
		return nil, errors.New("marker functionality not implemented")
	}

	// Find the end of the hostname(s) portion.
	end := bytes.IndexAny(line, "\t ")
	if end <= 0 {
		return nil, errors.New("bad format (insufficient fields)")
	}
	hosts := line[:end]
	keyBytes := line[end+1:]

	kh := &KnownHost{}

	// Check for hashed hostnames.
	if bytes.HasPrefix(hosts, []byte(sshHashPrefix)) {
		hosts = bytes.TrimPrefix(hosts, []byte(sshHashPrefix))
		// Hashed hostname format:
		//  <host>     = the hostname/address to be hashed
		//  <salt_b64> = base64(random 64 bits)
		//  <hash_b64> = base64(SHA1(<salt> <host>))
		//  <salt/hash pair> = '|1|' salt_b64 '|' hash_b64
		delim := bytes.Index(hosts, []byte(sshHashDelim))
		if delim <= 0 || delim >= len(hosts) {
			return nil, errors.New("bad hashed hostname format")
		}
		salt64 := hosts[:delim]
		hash64 := hosts[delim+1:]
		b64 := base64.StdEncoding
		kh.Salt = make([]byte, b64.DecodedLen(len(salt64)))
		kh.Hash = make([]byte, b64.DecodedLen(len(hash64)))
		if n, err := b64.Decode(kh.Salt, salt64); err != nil {
			return nil, err
		} else {
			kh.Salt = kh.Salt[:n]
		}
		if n, err := b64.Decode(kh.Hash, hash64); err != nil {
			return nil, err
		} else {
			kh.Hash = kh.Hash[:n]
		}
	} else {
		kh.Hostnames = strings.Split(string(hosts), ",")
	}

	// Finally, actually try to extract the key.
	key, _, _, _, err := ssh.ParseAuthorizedKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %v", err)
	}
	kh.Key = key

	return kh, nil
}
