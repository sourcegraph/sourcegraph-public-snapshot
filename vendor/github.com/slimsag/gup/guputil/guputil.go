package guputil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"math/big"

	"github.com/kr/binarydist"
)

// BundleInfo represents information about a gup patch bundle.
type BundleInfo struct {
	// Checksum of the new binary in the gup patch bundle.
	Checksum string

	// DecodedChecksum of the binary in the gup patch bundle.
	DecodedChecksum []byte

	// Signature is the ECDSA signature of the new binary in the gup patch
	// bundle.
	Signature ECDSASignature
}

func (b *BundleInfo) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "bundle info:\n")
	fmt.Fprintf(&buf, "  checksum:  %s (new file)\n", b.Checksum)

	rawSignature, _ := asn1.Marshal(b.Signature)
	sig := sha256.Sum256(rawSignature)
	fmt.Fprintf(&buf, "  signature: %s (SHA256 of ECDSA signature)\n", hex.EncodeToString(sig[:]))
	return buf.String()
}

// Diff computes a diff between old and new and writes out a gup patch bundle
// to patch. The ECDSA binary signature for new is written into the gup patch
// bundle using the provided private key.
func Diff(key *ecdsa.PrivateKey, old, new io.Reader, patch io.Writer) (*BundleInfo, error) {
	gw := gzip.NewWriter(patch)
	defer gw.Close()
	out := tar.NewWriter(gw)

	newData, err := ioutil.ReadAll(new)
	if err != nil {
		return nil, err
	}

	// Write the checksum of the new binary.
	encChecksum, decChecksum, err := computeChecksum(bytes.NewReader(newData))
	if err != nil {
		return nil, err
	}
	if err := out.WriteHeader(&tar.Header{
		Name: "checksum",
		Mode: 0600,
		Size: int64(len(encChecksum)),
	}); err != nil {
		return nil, err
	}
	if _, err := out.Write(encChecksum); err != nil {
		return nil, err
	}

	if old != nil {
		// Write the bsdiff patch file.
		var patchBuf bytes.Buffer // because we must provide size in header
		if err := binarydist.Diff(old, bytes.NewReader(newData), &patchBuf); err != nil {
			return nil, err
		}
		if err := out.WriteHeader(&tar.Header{
			Name: "bin.patch",
			Mode: 0600,
			Size: int64(patchBuf.Len()),
		}); err != nil {
			return nil, err
		}
		if _, err := out.Write(patchBuf.Bytes()); err != nil {
			return nil, err
		}
	} else {
		// Write the replacement binary file.
		if err := out.WriteHeader(&tar.Header{
			Name: "bin",
			Mode: 0600,
			Size: int64(len(newData)),
		}); err != nil {
			return nil, err
		}
		if _, err := out.Write(newData); err != nil {
			return nil, err
		}
	}

	// Write the new binary ECDSA signature.
	hash := sha256.New()
	if _, err := hash.Write(newData); err != nil {
		return nil, err
	}
	r, s, err := ecdsa.Sign(rand.Reader, key, hash.Sum(nil))
	if err != nil {
		return nil, err
	}
	signature := ECDSASignature{r, s}
	rawSignature, err := asn1.Marshal(signature)
	if err != nil {
		return nil, err
	}
	if err := out.WriteHeader(&tar.Header{
		Name: "bin.signature",
		Mode: 0600,
		Size: int64(len(rawSignature)),
	}); err != nil {
		return nil, err
	}
	if _, err := out.Write(rawSignature); err != nil {
		return nil, err
	}

	info := &BundleInfo{
		Checksum:        string(encChecksum),
		DecodedChecksum: decChecksum,
		Signature:       signature,
	}
	return info, out.Close()
}

// Patch applies the gup patch bundle, patch, to old, and writes the result to
// new. The provided public key is used to verify the ECDSA binary signature
// previously encoded into the gup patch bundle.
func Patch(pub *ecdsa.PublicKey, old io.Reader, new io.Writer, patch io.Reader) (*BundleInfo, error) {
	gzPatch, err := gzip.NewReader(patch)
	if err != nil {
		return nil, err
	}
	defer gzPatch.Close()
	in := tar.NewReader(gzPatch)

	// Iterate through the files in the archive.
	var (
		newBuf                                                  bytes.Buffer
		newDecChecksum                                          []byte
		havePatch, haveReplacement, haveChecksum, haveSignature bool

		info = &BundleInfo{}
	)
	for {
		hdr, err := in.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return nil, err
		}

		switch hdr.Name {
		case "bin.patch":
			havePatch = true
			err = binarydist.Patch(old, &newBuf, in)
			if err != nil {
				return nil, err
			}
			// TODO(slimsag): encoded form is not needed here
			_, newDecChecksum, err = computeChecksum(bytes.NewReader(newBuf.Bytes()))
			if err != nil {
				return nil, err
			}

		case "bin":
			haveReplacement = true
			_, err = io.Copy(&newBuf, in)
			if err != nil {
				return nil, err
			}
			// TODO(slimsag): encoded form is not needed here
			_, newDecChecksum, err = computeChecksum(bytes.NewReader(newBuf.Bytes()))
			if err != nil {
				return nil, err
			}

		case "checksum":
			haveChecksum = true
			encChecksum, decChecksum, err := decodeChecksum(in)
			if err != nil {
				return nil, err
			}
			info.Checksum = string(encChecksum)
			info.DecodedChecksum = decChecksum

		case "bin.signature":
			haveSignature = true
			raw, err := ioutil.ReadAll(in)
			if err != nil {
				return nil, err
			}
			if _, err := asn1.Unmarshal(raw, &info.Signature); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("gup: unexpected file in patch bundle %q", hdr.Name)
		}
	}
	if !havePatch && !haveReplacement {
		return nil, errors.New("gup: patch bundle contains no patch or replacement file")
	}
	if !haveChecksum || info.Checksum == "" || len(info.DecodedChecksum) == 0 {
		return nil, errors.New("gup: patch bundle is missing checksum file")
	}
	if !haveSignature {
		return nil, errors.New("gup: patch bundle is missing signature file")
	}

	if !bytes.Equal(newDecChecksum, info.DecodedChecksum) {
		return nil, errors.New("gup: checksum verification failed")
	}
	if !ecdsa.Verify(pub, newDecChecksum, info.Signature.R, info.Signature.S) {
		return nil, errors.New("gup: ECDSA signature verification failed")
	}

	_, err = io.Copy(new, &newBuf)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func Checksum(r io.Reader) (string, error) {
	checksum := sha256.New()
	if _, err := io.Copy(checksum, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(checksum.Sum(nil)), nil
}

func computeChecksum(r io.Reader) ([]byte, []byte, error) {
	checksum := sha256.New()
	if _, err := io.Copy(checksum, r); err != nil {
		return nil, nil, err
	}
	ck := checksum.Sum(nil)
	encoded := make([]byte, hex.EncodedLen(len(ck)))
	hex.Encode(encoded, ck)
	return encoded, ck, nil
}

func decodeChecksum(r io.Reader) ([]byte, []byte, error) {
	encoded, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	decoded := make([]byte, hex.DecodedLen(len(encoded)))
	if _, err := hex.Decode(decoded, encoded); err != nil {
		return nil, nil, err
	}
	return encoded, decoded, nil
}

type ECDSASignature struct {
	R, S *big.Int
}

// ParsePublicKey parses a PEM encoded public key.
func ParsePublicKey(key []byte) (pubKey *ecdsa.PublicKey, err error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("ParsePublicKey: failed to parse public key")
	}
	if block.Type != "PUBLIC KEY" {
		return nil, errors.New("ParsePublicKey: expected public key to have block type PUBLIC KEY")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	v, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("ParsePublicKey: expected ECDSA publi key, found %T", pub)
	}
	return v, nil
}

// Index represents a index.json file.
type Index struct {
	// Tags is a map of all tags (or "release channels") to their respective
	// versions.
	Tags map[string]*IndexVersions
}

// IndexVersions represents versions found in a index.json file for a specific tag.
type IndexVersions struct {
	// List is a list of all versions available, in order of most recent (index
	// zero) to oldest (last index).
	List []IndexVersion
}

// IndexVersion represents a single version found in a index.json file.
//
// The first version in a chain will always have an empty From value equal to
// the To value and will be a Replacement binary.
type IndexVersion struct {
	// From and To represent the SHA256 hash of the old (from) file and the new
	// (to) file. For example, if a user's binary is currently SHA256="foo"
	// then any version where From="foo" would be the next update for that
	// user.
	From, To string

	// Replacement indicates that the version is a full binary replacement,
	// i.e. not an incremental patch.
	Replacement bool
}

// FindNextVersion returns the next logical version to upgrade to when coming
// from the given SHA256. If nil is returned, there is no next version.
func (v IndexVersions) FindNextVersion(from string) (*IndexVersion, int) {
	for i := len(v.List) - 1; i >= 0; i-- {
		v := v.List[i]
		if v.To == from {
			// There are two cases where this occurs, and in both cases we
			// never want to continue further.
			//
			// In the first case, this is the first update in the chain
			// (`i == 0`), e.g.:
			//
			// 	0: {From: v1, To: v1, Replacement: true}
			//
			// And we never want to update from the same version to the same
			// version.
			//
			// In the second case, we have a chain like so:
			//
			// 	0: {From: v1, To: v1, Replacement: true}
			// 	1: {From: v1, To: v2, Replacement: false}
			// 	2: {From: v2, To: v3, Replacement: false}
			//
			// And we never want to look any further past the `i`th element,
			// e.g. for `from=v3`, we do not want to look at index 1 or 0. For
			// `from=v2`, we want to look at index 2 but not index 1 or 0. i.e.
			// any version lesser than ours.
			return nil, 0
		}

		// If any recent version is a full replacement, always choose that as
		// it is the shortest link to the latest version.
		if v.Replacement {
			return &v, i
		}
		if v.From == from {
			return &v, i
		}
	}
	return nil, 0
}

// FindCurrentVersion finds the current version for x; i.e. the version whose
// To specifies x.
func (v IndexVersions) FindCurrentVersion(x string) (*IndexVersion, int) {
	for i := len(v.List) - 1; i >= 0; i-- {
		v := v.List[i]
		if v.To == x {
			return &v, i
		}
	}
	return nil, 0
}

// UpdateFilename returns the update filename for the specified tag and version.
func UpdateFilename(tag string, versionIndex int) string {
	return fmt.Sprintf("%s-%d.tgz", tag, versionIndex)
}

// Tag creates a tag given the user tag, GOOS, and GOARCH.
func Tag(tag, goos, goarch string) string {
	return fmt.Sprintf("%s-%s-%s", tag, goos, goarch)
}

// ExpandTag expands a tag for usage in an index.json file.
func ExpandTag(tag string) string {
	return Tag(tag, build.Default.GOOS, build.Default.GOARCH)
}
