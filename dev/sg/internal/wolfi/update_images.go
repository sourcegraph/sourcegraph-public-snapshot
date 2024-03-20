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


func GetAllImages() (imageNames []string, err error) {
	// Iterate over *.yaml files in wolfi-images/
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return nil, err
	}
	imageDir := filepath.Join(repoRoot, baseImageDir)
	files, err := os.ReadDir(imageDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		imageName := strings.TrimSuffix(file.Name(), ".yaml")
		imageNames = append(imageNames, imageName)
	}

	return imageNames, nil
}

// UpdateAllImages runs UpdateImage for all images in the baseImageDir
func UpdateAllImages(ctx *cli.Context, enableLocalPackageRepo bool) error {
	imageNames, err := GetAllImages()
	if err != nil {
		return err
	}

	for _, imageName := range imageNames {
		bc, err := SetupBaseImageBuild(imageName, PackageRepoConfig{})
		if err != nil {
			return err
		}

		bc.UpdateImage(ctx, enableLocalPackageRepo)
	}

	return nil
}

func (bc BaseImageConfig) UpdateImage(_ *cli.Context, enableLocalPackageRepo bool) error {

	var extraRepo, extraKey string
	if enableLocalPackageRepo {
		// Currently not implemented as rules_apko doesn't support local filesystem repos
	}

	// Update lockfile
	std.Out.WriteLine(output.Linef("🗝️ ", output.StylePending, fmt.Sprintf("Updating apko lockfile for %s", bc.ImageName)))
	if err := bc.ApkoLock(extraRepo, extraKey); err != nil {
		return err
	}

	return nil
}

func (bc BaseImageConfig) ApkoLock(extraRepo string, extraKey string) error {
	localImageConfigPath := strings.TrimPrefix(bc.ImageConfigPath, bc.ImageConfigDir+"/")

	apkoArgs := []string{"run", "@rules_apko//apko", "lock", "--", localImageConfigPath}

	apkoFlags := []string{}
	if extraRepo != "" {
		apkoFlags = append(apkoFlags, "--repository-append", extraRepo)
	}
	if extraKey != "" {
		apkoFlags = append(apkoFlags, "--keyring-append", extraKey)
	}

	cmd := exec.Command("bazel", append(apkoArgs, apkoFlags...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = bc.ImageConfigDir
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to build base image")
	}

	// Update hash in lockfile
	_, err = bc.CheckApkoLockHash(true)
	if err != nil {
		return err
	}

	return nil
}

func CheckApkoLockHashes(imageNames []string) (allImagesMatch bool, mismatchedImages []string, err error) {
	if len(imageNames) == 0 {
		imageNames, err = GetAllImages()
		if err != nil {
			return false, nil, err
		}
	}

	for _, imageName := range imageNames {
		bc, err := SetupBaseImageBuild(imageName, PackageRepoConfig{})
		if err != nil {
			return false, nil, err
		}

		imageSynced, err := bc.CheckApkoLockHash(false)
		if err != nil {
			return false, nil, err
		}

		if !imageSynced {
			allImagesMatch = false
			mismatchedImages = append(mismatchedImages, imageName)
		}
	}

	return allImagesMatch, mismatchedImages, nil
}

// CheckApkoLockHash checks whether the hash of an image's YAML file matches the hash stored in the corresponding lockfile
// This allows us to detect changes to the YAML file and re-run apko lock if necessary
func (bc BaseImageConfig) CheckApkoLockHash(update bool) (isMatch bool, err error) {
	apkoConfig, err := os.ReadFile(bc.ImageConfigPath)
	if err != nil {
		return false, err
	}

	apkoConfigHash := sha256.Sum256([]byte(apkoConfig))
	apkoConfigHashHex := hex.EncodeToString(apkoConfigHash[:])
	fmt.Printf("apkoConfigHashHex: %s\n", apkoConfigHashHex) // TODO: Remove

	// Now we have the hash, we need to add it to the json
	// TODO: What about the case where the lockfile doesn't exist? That's a valid false, not an error
	imageLock, err := os.ReadFile(bc.LockfilePath)
	if err != nil {
		return false, err
	}

	var imageLockData map[string]interface{}
	err = json.Unmarshal(imageLock, &imageLockData)
	if err != nil {
		return false, err
	}

	if val, exists := imageLockData["configHash"]; exists {
		fmt.Println("configHash before:", val) // TODO: Remove
		if val == apkoConfigHashHex {
			isMatch = true
		}
	} else {
		fmt.Println("configHash key not found") // TODO: Remove
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
	err = os.WriteFile(bc.LockfilePath, updatedFile, 0644)
	if err != nil {
		return false, err
	}

	return isMatch, nil
}
