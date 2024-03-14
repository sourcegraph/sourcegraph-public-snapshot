package wolfi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func UpdateImages(_ *cli.Context, updateImageName string) error {
	// Loop over all images and run apko lock
	if updateImageName != "" {
		if !strings.HasSuffix(updateImageName, ".yaml") {
			updateImageName = updateImageName + ".yaml"
		}
	}

	// Iterate over *.yaml files in wolfi-images/
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	imageDir := filepath.Join(repoRoot, "wolfi-images")
	files, err := os.ReadDir(imageDir)
	if err != nil {
		return err
	}

	var updatedImage bool
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		if updateImageName != "" && file.Name() != updateImageName {
			continue
		}

		// Update lockfile
		std.Out.WriteLine(output.Linef("🗝️ ", output.StylePending, fmt.Sprintf("Updating apko lockfile for %s", file.Name())))
		if err = ApkoLock(file.Name(), imageDir, "", ""); err != nil {
			return err
		}
		updatedImage = true
	}

	if updateImageName != "" && !updatedImage {
		return errors.New(fmt.Sprintf("no such image '%s'", updateImageName))
	}

	return nil
}

func ApkoLock(imageFilename string, imageDir string, extraRepo string, extraKey string) error {

	apkoFlags := []string{"lock", imageFilename}
	if extraRepo != "" {
		apkoFlags = append(apkoFlags, "--repository-append", extraRepo)
	}
	if extraKey != "" {
		apkoFlags = append(apkoFlags, "--keyring-append", extraKey)
	}

	// TODO: Replace with bazel command
	cmd := exec.Command("apko", apkoFlags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = imageDir
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to build base image")
	}

	// Update hash in lockfile
	// TODO: Improve this handling
	imageName := strings.TrimSuffix(imageFilename, ".yaml")
	_, err = CheckApkoLockHash(imageName, true)
	if err != nil {
		return err
	}

	return nil
}

func getImageConfigFilename(imageName string) (string, error) {
	return getApkoConfigFile(imageName, ".yaml")
}
func getImageLockFilename(imageName string) (string, error) {
	return getApkoConfigFile(imageName, ".lock.json")
}
func getApkoConfigFile(imageName string, suffix string) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}
	imageConfigDir := "wolfi-images"
	imageConfigFile := filepath.Join(repoRoot, imageConfigDir, imageName+suffix)

	return imageConfigFile, nil
}

// CheckApkoLockHash checks whether the hash of an image's YAML file matches the hash stored in the corresponding lockfile
// This allows us to detect changes to the YAML file and re-run apko lock if necessary
func CheckApkoLockHash(imageName string, update bool) (isMatch bool, err error) {
	apkoConfigFile, err := getImageConfigFilename(imageName)
	if err != nil {
		return false, err
	}

	apkoConfig, err := os.ReadFile(apkoConfigFile)
	if err != nil {
		return false, err
	}

	apkoConfigHash := sha256.Sum256([]byte(apkoConfig))
	apkoConfigHashHex := hex.EncodeToString(apkoConfigHash[:])
	fmt.Printf("apkoConfigHashHex: %s\n", apkoConfigHashHex)

	// Now we have the hash, we need to add it to the json
	imageLockFile, err := getImageLockFilename(imageName)
	imageLock, err := os.ReadFile(imageLockFile)
	if err != nil {
		return false, err
	}

	var imageLockData map[string]interface{}
	err = json.Unmarshal(imageLock, &imageLockData)
	if err != nil {
		return false, err
	}

	if val, exists := imageLockData["configHash"]; exists {
		fmt.Println("configHash before:", val)
		if val == apkoConfigHashHex {
			isMatch = true
		}
	} else {
		fmt.Println("configHash key not found")
	}

	if !update {
		return isMatch, nil
	}

	imageLockData["configHash"] = apkoConfigHashHex

	// Marshal the map back to json
	updatedFile, err := json.MarshalIndent(imageLockData, "", "  ")
	if err != nil {
		return false, err
	}

	// Write the updated json back to the file
	err = os.WriteFile(imageLockFile, updatedFile, 0644)
	if err != nil {
		return false, err
	}

	return isMatch, nil
}
