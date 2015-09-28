package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"sourcegraph.com/sourcegraph/sourcegraph/dev/release"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/sgxcmd"
)

func init() {
	_, err := CLI.AddCommand("release",
		"release a new version to all users",
		"The release command releases a new Sourcegraph version by building, packaging, and uploading it.",
		&releaseCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

type ReleaseCmd struct {
	SkipPackage      bool `long:"skip-package" description:"skip package step (assumes it has already been run)"`
	SkipDistPackage  bool `long:"skip-dist-package" description:"skip create deb and rpm step (assumes it has already been run)"`
	InspectArtifacts bool `long:"inspect-artifacts" description:"avoids upload, but puts all artifacts in ./selfupdate"`

	S3Dir string `long:"s3-dir" description:"S3 base directory to upload release to (default: src)"`

	PackageCmd
}

var releaseCmd ReleaseCmd

func (c *ReleaseCmd) Execute(args []string) error {
	if !c.SkipPackage {
		if err := c.PackageCmd.Execute(nil); err != nil {
			return err
		}
	}
	if !c.SkipDistPackage {
		cmd := exec.Command("make", "package", "VERSION="+c.Args.Version)
		cmd.Dir = "./package"
		if err := execCmd(cmd); err != nil {
			return err
		}
	}

	var selfupdateDir string
	if c.InspectArtifacts {
		selfupdateDir = "selfupdate"
	} else {
		var err error
		selfupdateDir, err = ioutil.TempDir("", "selfupdate")
		if err != nil {
			return err
		}
		defer os.RemoveAll(selfupdateDir)
	}

	const releaseDir = "release"
	if err := execCmd(exec.Command("go-selfupdate", "-o="+selfupdateDir, "-cmd="+sgxcmd.Name, filepath.Join(releaseDir, c.Args.Version), c.Args.Version)); err != nil {
		return err
	}
	distDir := "package/dist/" + c.Args.Version
	if err := execCmd(exec.Command("cp", distDir+"/src.deb", distDir+"/src.rpm", selfupdateDir+"/"+c.Args.Version+"/linux-amd64/")); err != nil {
		return err
	}
	if err := execCmd(exec.Command("cp", distDir+"/src-docker.deb", selfupdateDir+"/"+c.Args.Version+"/linux-amd64/")); err != nil {
		return err
	}

	if c.InspectArtifacts {
		return nil
	}

	if c.S3Dir == "" {
		c.S3Dir = release.S3Dir
	}

	syncCmd := exec.Command(
		"aws", "s3", "sync",
		"--acl", "public-read",
		selfupdateDir,
		"s3://"+release.S3Bucket+"/"+c.S3Dir,
	)
	if err := execCmd(syncCmd); err != nil {
		return err
	}

	return nil
}
