package lock

import (
	"encoding/json"
	"fmt"
	"os"
)

type Lock struct {
	Version  string       `json:"version"`
	Contents LockContents `json:"contents"`
}

type LockContents struct {
	Keyrings     []LockKeyring `json:"keyring"`
	Repositories []LockRepo    `json:"repositories"`
	// Packages in order of installation -> for a single architecture.
	Packages []LockPkg `json:"packages"`
}

type LockPkg struct {
	Name         string                  `json:"name"`
	URL          string                  `json:"url"`
	Version      string                  `json:"version"`
	Architecture string                  `json:"architecture"`
	Signature    LockPkgRangeAndChecksum `json:"signature"`
	Control      LockPkgRangeAndChecksum `json:"control"`
	Data         LockPkgRangeAndChecksum `json:"data"`
	// Checksum is APK-style: 'Q1' prefixed SHA1 hash of the second gzip stream (control stream) in the package.
	// For data-consistency checks use Signature, Control & Data above.
	// Populated since Apko 0.13.
	Checksum string `json:"checksum"`
}
type LockPkgRangeAndChecksum struct {
	Range    string `json:"range"`
	Checksum string `json:"checksum"`
}

type LockRepo struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Architecture string `json:"architecture"`
}

type LockKeyring struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func FromFile(lockFile string) (Lock, error) {
	payload, err := os.ReadFile(lockFile)
	if err != nil {
		return Lock{}, fmt.Errorf("failed to load lockfile: %w", err)
	}
	var lock Lock
	err = json.Unmarshal(payload, &lock)
	return lock, err
}

func (lock Lock) SaveToFile(lockFile string) error {
	jsonb, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshall json: %w", err)
	}
	// Github and pre-commit checks (like end-of-file-fixer) are expecting ASCII files
	// to end with a newline that marshal is not providing.
	jsonb = append(jsonb, '\n')
	return os.WriteFile(lockFile, jsonb, os.ModePerm)
}
