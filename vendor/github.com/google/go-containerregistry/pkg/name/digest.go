// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package name

import (
	// nolint: depguard
	_ "crypto/sha256" // Recommended by go-digest.
	"strings"

	"github.com/opencontainers/go-digest"
)

const digestDelim = "@"

// Digest stores a digest name in a structured form.
type Digest struct {
	Repository
	digest   string
	original string
}

// Ensure Digest implements Reference
var _ Reference = (*Digest)(nil)

// Context implements Reference.
func (d Digest) Context() Repository {
	return d.Repository
}

// Identifier implements Reference.
func (d Digest) Identifier() string {
	return d.DigestStr()
}

// DigestStr returns the digest component of the Digest.
func (d Digest) DigestStr() string {
	return d.digest
}

// Name returns the name from which the Digest was derived.
func (d Digest) Name() string {
	return d.Repository.Name() + digestDelim + d.DigestStr()
}

// String returns the original input string.
func (d Digest) String() string {
	return d.original
}

// NewDigest returns a new Digest representing the given name.
func NewDigest(name string, opts ...Option) (Digest, error) {
	// Split on "@"
	parts := strings.Split(name, digestDelim)
	if len(parts) != 2 {
		return Digest{}, newErrBadName("a digest must contain exactly one '@' separator (e.g. registry/repository@digest) saw: %s", name)
	}
	base := parts[0]
	dig := parts[1]
	prefix := digest.Canonical.String() + ":"
	if !strings.HasPrefix(dig, prefix) {
		return Digest{}, newErrBadName("unsupported digest algorithm: %s", dig)
	}
	hex := strings.TrimPrefix(dig, prefix)
	if err := digest.Canonical.Validate(hex); err != nil {
		return Digest{}, err
	}

	tag, err := NewTag(base, opts...)
	if err == nil {
		base = tag.Repository.Name()
	}

	repo, err := NewRepository(base, opts...)
	if err != nil {
		return Digest{}, err
	}
	return Digest{
		Repository: repo,
		digest:     dig,
		original:   name,
	}, nil
}
