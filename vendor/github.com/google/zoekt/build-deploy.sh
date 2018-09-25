#!/bin/sh

# this script packages up all the binaries, and a script (deploy.sh)
# to twiddle with the server and the binaries

set -ex

# Put the date first so we can sort.
if [[ -z "$VERSION" ]]; then
  VERSION=$(date --iso-8601=minutes | tr -d ':' | sed 's|\+.*$||')
  if [[ -d .git ]]; then
    VERSION=${VERSION}-$(git show --pretty=format:%h -q)
  fi
fi

set -u

out=zoekt-${VERSION}
mkdir -p ${out}

for d in cmd/*
do
  go build -tags netgo -ldflags "-X github.com/google/zoekt.Version=$VERSION"  -o ${out}/$(basename $d) github.com/google/zoekt/$d
done

cat <<EOF > ${out}/deploy.sh
#!/bin/bash

echo "Set the following in the environment."
echo ""
echo '  export PATH="'$PWD'/bin:$PATH'
echo ""

set -eux

# Allow sandbox to create NS's
sudo sh -c 'echo 1 > /proc/sys/kernel/unprivileged_userns_clone'

# we mmap the entire index, but typically only want the file contents.
sudo sh -c 'echo 1 >/proc/sys/vm/overcommit_memory'

# allow bind to 80 and 443
sudo setcap 'cap_net_bind_service=+ep' bin/zoekt-webserver

EOF

chmod 755 ${out}/*

tar --owner=root --group=root -czf zoekt-deploy-${VERSION}.tar.gz ${out}/*

rm -rf ${out}
