set -ex
git tag $VERSION -a -m "release v$VERSION"
git tag latest -f -a -m "release v$VERSION"
git push -f --tags
