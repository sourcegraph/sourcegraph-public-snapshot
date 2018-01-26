# log15's release strategy

log15 uses gopkg.in to manage versioning releases so that consumers who don't vendor dependencies can rely upon a stable API.

## Master

Master is considered to have no API stability guarantee, so merging new code that passes tests into master is always okay.

## Releasing a new API-compatible version

The process to release a new API-compatible version is described below. For the purposes of this example, we'll assume you're trying to release a new version of v2

1. `git checkout v2`
1. `git merge master`
1. Audit the code for any imports of sub-packages. Modify any import references from `github.com/inconshrevealbe/log15/<pkg>` -> `gopkg.in/inconshreveable/log15.v2/<pkg>`
1. `git commit`
1. `git tag`, find the latest tag of the style v2.X.
1. `git tag v2.X+1` If the last version was v2.6, you would run `git tag v2.7`
1. `git push --tags git@github.com:inconshreveable/log15.git v2`
