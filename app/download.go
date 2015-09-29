package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/dev/release"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

var downloadBaseURL = "https://" + release.S3Bucket + ".s3.amazonaws.com/" + sgxcmd.Name + "/"

func serveDownload(w http.ResponseWriter, r *http.Request) error {
	prefix := router.Rel.URLTo(router.Download, "Suffix", "")
	target := strings.TrimPrefix(r.URL.Path, prefix.Path)

	// If the user is requesting the "latest" binary then we fill in the version
	// for them as part of the redirect.
	if s := strings.Split(target, "/"); len(s) >= 1 && s[0] == "latest" {
		// Determine latest version.
		resp, err := http.Get(downloadBaseURL + "linux-amd64/src.json")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var data struct {
			Version string
			Sha256  string
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}

		// Reform target URL by swapping out "latest" with the version.
		target = path.Join(data.Version, path.Join(s[1:]...))
	}
	http.Redirect(w, r, downloadBaseURL+target, http.StatusSeeOther)
	return nil
}

func serveDownloadInstall(w http.ResponseWriter, r *http.Request) error {
	// Write the bash script.
	fmt.Fprintf(w, `#!/bin/bash

# This bash script is meant to be piped directly into bash:
#
# via cURL:
#
#  curl -sSL https://sourcegraph.com/.download/install.sh | bash
#
# via wget:
#
#  wget -O - https://sourcegraph.com/.download/install.sh | bash
#
# It automatically performs the installation process of Sourcegraph onto the
# system, by simply detecting the OS and installing the relevant package. In
# this way, uninstallation can be performed simply via your system's normal
# package manager.
#
# All your Sourcegraph data (repos, etc) is stored in the ~/.sourcegraph
# directory, and your OAuth tokens are stored in the ~/.src-auth file.
#
# Visit sourcegraph.com for more information. You can also reach us at
# support@sourcegraph.com should you have any questions, comments or concerns.
# We'd love to hear from you!

set -e

on_error() {
	set +x # echo off
	echo
	echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
	echo "!! ERROR! One or more of the commands above failed to run!                    !!"
	echo "!! -> Please contact support@sourcegraph.com and include the above output!    !!"
	echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
	exit 1
}

# have_git tells if the git command is installed or not.
have_git() {
	trap '' ERR # unset trap
	set +e # unset exit on error

	git --version 2>&1 >/dev/null
	ok=$?

	set -e # set exit on error
	trap 'on_error' ERR # set trap

	if [ $ok -eq 0 ]; then
		return 0
	else
		return 1
	fi
}

do_install() {
	trap 'on_error' ERR

	# Create tmp directory, this works on OS X and Linux (see http://unix.stackexchange.com/a/84980).
	download_dir=$(mktemp -d 2>/dev/null || mktemp -d -t 'sourcegraph')

	# Detect the OS using the pattern described at http://stackoverflow.com/a/17072017
	if [ "$(uname)" == "Darwin" ]; then
		# OS X
		set -x # echo on

		# Install git if it's not already installed.
		if ! have_git; then
			set +x; echo "Installing git..."; set -x
			sudo brew install git
		fi

		# OS X needs /usr/local/bin to be created because on default installations
		# it is not already (mostly of the time it is created by homebrew, but we
		# don't want to require that).
		sudo mkdir -p /usr/local/bin

		# OS X doesn't always have /usr/local/bin on the $PATH so we add an entry
		# for it here only if one does not yet exist.
		echo $PATH | grep /usr/local/bin &> /dev/null || echo export PATH='/usr/local/bin:$PATH' >> ~/.bash_profile

		# Download the file into the tmp directory and unzip it.
		pushd $download_dir
		echo
		set -x # echo on
		curl -O -L https://sourcegraph.com/.download/latest/darwin-amd64/src.gz
		gunzip src.gz
		chmod +x src
		sudo mv src /usr/local/bin
		popd

		set +x # echo off
		echo
		echo "********************************************************************************"
		echo "** Success! Sourcegraph has been installed!                                   **"
		echo "** -> Run 'src serve' to start Sourcegraph!                                   **"
		echo "********************************************************************************"

	elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
		# Linux
		set -x # echo on

		# Determine if system is rpm or deb based, see:
		#
		# https://ask.fedoraproject.org/en/question/49738/how-to-check-if-system-is-rpm-or-debian-based/
		#
		trap '' ERR # unset trap
		set +e # unset exit on error

		/usr/bin/rpm -q -f /usr/bin/rpm >/dev/null 2>&1
		rpm_based=$?

		set -e # set exit on error
		trap 'on_error' ERR # set trap

		# Download the file into the tmp directory and install using dpkg or yum.
		pushd $download_dir
		if [ $rpm_based -eq 0 ]; then
			# Install git if it's not already installed.
			if ! have_git; then
				set +x; echo "Installing git..."; set -x
				sudo yum -y install git
			fi

			echo "Installing the rpm package"
			wget https://sourcegraph.com/.download/latest/linux-amd64/src.rpm
			sudo yum install src.rpm
		else
			# Install git if it's not already installed.
			if ! have_git; then
				set +x; echo "Installing git..."; set -x
				sudo apt-get install -y git
			fi

			echo "Installing the deb package"
			wget https://sourcegraph.com/.download/latest/linux-amd64/src.deb
			sudo dpkg -i src.deb
		fi
		popd

		set +x # echo off
		echo
		echo "********************************************************************************"
		echo "** Success! Sourcegraph has been installed!                                   **"
		echo "** -> Visit http://localhost:3000 to use Sourcegraph!                         **"
		echo "********************************************************************************"
	fi
}

# Just as many other install scripts do, we wrap everything in a function here
# as it is possible to get only half the file during 'curl | bash'.
do_install
`)
	return nil
}
