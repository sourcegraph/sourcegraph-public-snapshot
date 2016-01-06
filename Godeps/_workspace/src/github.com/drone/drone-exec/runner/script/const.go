package script

// entrypoint is the default bash entrypoint command
// slice, used to execute a build script in string format.
var entrypoint = []string{"/bin/sh", "-e", "-c"}

// setupScript is a helper script this is added
// to the build to ensure a minimum set of environment
// variables are set correctly.
const setupScript = `
[ -z "$HOME"  ] && export HOME="/root"
[ -z "$SHELL" ] && export SHELL="/bin/sh"

export GOBIN=/drone/bin
export GOPATH=/drone
export PATH=$PATH:$GOBIN

set -e
`

const teardownScript = `
rm -rf $HOME/.netrc
rm -rf $HOME/.ssh/id_rsa
`

// netrcScript is a helper script that is added to
// the build script to enable cloning private git
// repositories of http.
const netrcScript = `
cat <<EOF >> $HOME/.netrc
machine %s
login %s
password %s

EOF
chmod 0600 $HOME/.netrc
`

// keyScript is a helper script that is added to
// the build script to add the id_rsa key to clone
// private repositories.
const keyScript = `
mkdir -p -m 0700 $HOME/.ssh
cat <<EOF > $HOME/.ssh/id_rsa
%s
EOF
chmod 0600 $HOME/.ssh/id_rsa
`

// keyConfScript is a helper script that is added
// to the build script to ensure that git clones don't
// fail due to strict host key checking prompt.
const keyConfScript = `
cat <<EOF > $HOME/.ssh/config
StrictHostKeyChecking no
EOF
`

// forceYesScript is a helper script that is added
// to the build script to ensure apt-get installs
// don't prompt the user to accept changes.
const forceYesScript = `
if [ "$(id -u)" = "0" ]; then
mkdir -p /etc/apt/apt.conf.d
cat <<EOF > /etc/apt/apt.conf.d/90forceyes
APT::Get::Assume-Yes "true";APT::Get::force-yes "true";
EOF
fi
`

// traceScript is a helper script that is added
// to the build script to trace a command.
const traceScript = `
echo %s | base64 -d
%s
`
