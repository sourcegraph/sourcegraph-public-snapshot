package gup

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	update "github.com/inconshreveable/go-update"
	"github.com/kardianos/osext"

	"github.com/slimsag/gup/guputil"
)

var Config = struct {
	// UpdateURL is a template URL that is contacted to look for both the
	// index.json file, as well as all update tarballs. It must contain "$GUP"
	// as the file portion of the URL.
	UpdateURL string

	// PublicKey is the ECDSA public key to use for verifying the binary
	// against the signature file in a gup bundle.
	PublicKey string

	// Tag, which defaults to "main", specifies the tag to use when looking for
	// updates. If the tag is invalid (i.e. not in index.json) updates will
	// fail.
	Tag string

	// CheckInterval is the interval at which updates are checked for while the
	// program is running. Zero signals to not check for updates while the
	// program is running. The default is one hour.
	CheckInterval time.Duration

	// Client is the HTTP client to use when fetching updates. By default, it
	// has a 5s timeout.
	Client *http.Client
}{
	CheckInterval: 1 * time.Hour,
	Tag:           "main",
	Client:        &http.Client{Timeout: 5 * time.Second},
}

// UpdateAvailable is a channel that users can read from to get a signal for
// when an update is available.
var UpdateAvailable = make(chan bool, 1)

// Update checks if an update is available, and if it is, applies it. If
// updating is attempted but fails, an error is returned. If no update is
// available, Update simply returns (after making the HTTP request).
func Update() (bool, error) {
	tag := guputil.ExpandTag(Config.Tag)
	idx, haveUpdate := checkNow(tag)
	if !haveUpdate {
		return false, nil // no update available
	}

	// Read the old (current) executable.
	exePath, err := osext.Executable()
	if err != nil {
		return false, err
	}
	old, err := ioutil.ReadFile(exePath)
	if err != nil {
		return false, err
	}

	// Patching to latest.
	var newBuf *bytes.Buffer
	for {
		oldChecksum, err := guputil.Checksum(bytes.NewReader(old))
		if err != nil {
			return false, err
		}
		version, versionIndex := idx.Tags[tag].FindNextVersion(oldChecksum)
		if version == nil {
			break // at latest version
		}

		// Fetch the update file.
		updateFile := guputil.UpdateFilename(tag, versionIndex)
		updateURL := strings.Replace(Config.UpdateURL, "$GUP", updateFile, -1)
		resp, err := http.Get(updateURL)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		// Apply the update in-memory.
		newBuf = new(bytes.Buffer)
		_, err = guputil.Patch(pub, bytes.NewReader(old), newBuf, resp.Body)
		if err != nil {
			return false, err
		}
		old = newBuf.Bytes()

		desc := "incremental"
		if version.Replacement {
			desc = "replacement"
		}
		fmt.Fprintf(os.Stderr, "applied %s update %s\n", desc, updateFile)
	}

	// Apply the update to our binary itself in-place. We delegate to go-update
	// for this, specifying no patcher (i.e. a binary replacement).
	return true, update.Apply(newBuf, update.Options{
		Patcher: nil,
	})
}

// Start starts gup. If the configuration is invalid, a panic occurs.
func Start() {
	// Configuration verification.
	if Config.UpdateURL == "" {
		panic("gup: gup.Config.UpdateURL must be set")
	}
	if !strings.Contains(Config.UpdateURL, "$GUP") {
		panic("gup: gup.Config.UpdateURL must contain $GUP in the URL")
	}
	if Config.PublicKey == "" {
		panic("gup: gup.Config.PublicKey must be set")
	}
	var err error
	pub, err = guputil.ParsePublicKey([]byte(Config.PublicKey))
	if err != nil {
		panic(fmt.Sprintf("gup: gup.Config.PublicKey error: %v", err))
	}

	// Background update checking.
	if Config.CheckInterval == time.Duration(0) {
		return
	}
	tag := guputil.ExpandTag(Config.Tag)
	go func() {
		if _, update := checkNow(tag); update {
			broadcastUpdate()
		}
		t := time.Tick(Config.CheckInterval)
		for {
			select {
			case <-t:
				if _, update := checkNow(tag); update {
					broadcastUpdate()
				}
			}
		}
	}()
}

var cachedHash string

func hashSelf() (string, error) {
	if cachedHash != "" {
		return cachedHash, nil
	}

	path, err := osext.Executable()
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return guputil.Checksum(f)
}

var curVersionWarn sync.Once

// CheckNow checks for updates immediately, and returns true if one is available.
func CheckNow() bool {
	_, u := checkNow(guputil.ExpandTag(Config.Tag))
	return u
}

// checkNow checks for updates immediately, and returns true if one is available.
func checkNow(tag string) (*guputil.Index, bool) {
	// TODO(slimsag): configurable logging
	indexURL := strings.Replace(Config.UpdateURL, "$GUP", "index.json", -1)
	resp, err := http.Get(indexURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "checking for updates: %v (url %s)\n", err, indexURL)
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "checking for updates: got %s (url %s)\n", resp.Status, indexURL)
		return nil, false
	}

	var idx guputil.Index
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		fmt.Fprintf(os.Stderr, "checking for updates: %v (url %s)\n", err, indexURL)
		return nil, false
	}

	from, err := hashSelf()
	if err != nil {
		fmt.Fprintf(os.Stderr, "checking for updates: %v (url %s)\n", err, indexURL)
		return nil, false
	}

	if _, ok := idx.Tags[tag]; !ok {
		return nil, false // tag doesn't exist
	}
	current, _ := idx.Tags[tag].FindCurrentVersion(from)
	if current == nil {
		curVersionWarn.Do(func() {
			fmt.Printf("checking for updates: warning: current version %q not in index\n", from)
		})
		return nil, false
	}

	version, _ := idx.Tags[tag].FindNextVersion(from)
	return &idx, version != nil
}

// broadcastUpdate performs a non-blocking send to the UpdateAvailable channel.
func broadcastUpdate() {
	select {
	case UpdateAvailable <- true:
	default:
	}
}

var pub *ecdsa.PublicKey

type patcherFunc func(old io.Reader, new io.Writer, patch io.Reader) error

func (p patcherFunc) Patch(old io.Reader, new io.Writer, patch io.Reader) error {
	return p(old, new, patch)
}
