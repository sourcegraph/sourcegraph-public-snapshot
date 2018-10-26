# Sourcegraph development process

1.  Configure your repository with two remotes, `oss` and `ent`, by running the following:

```bash
enterprise/dev/init-repo.sh
```

After running this script, the following should hold:

- `oss` should point to `https://github.com/sourcegraph/sourcegraph`.
- `ent` should point to `https://github.com/sourcegraph/enterprise`.
- There should be no `origin` remote.

2.  Decide whether you will branch off the open-source (`oss`) or enterprise (`ent`) repo. When in
    doubt, prefer `oss`.

## Developing off `oss`

1.  Run `dev/launch.sh` from the root of this repository.
1.  Push your branch up to `oss` and open a PR.
1.  Wait for CI to pass, resolve any merge conflicts.
1.  Merge the PR into `oss` `master`. This will trigger another PR for you in `ent` that contains the
    commits you just merged. This PR will be automatically merged if `ent` CI passes and there are no
    merge conflicts.
1.  If the `ent` PR cannot be automatically merged, update the PR to resolve any CI errors or merge
    conflicts.
    - If you make any changes to OSS code in the `ent` repository, you are responsible for ensuring
      these changes make it into `oss`.

## Developing off `ent`

If a substantial subset of your change can be made as an independent change to `oss`, prefer to do
that.

1.  Run `dev/start.sh` from `enterprise` directory.
1.  Push your branch up to `ent` and open a PR.
1.  Wait for CI to pass, resolve any merge conflicts.
1.  Merge the PR into `ent` `master`.
1.  Cherry-pick your changes onto `oss/master` using `dev/prune-pick.sh`.
1.  If there are no conflicts, push directly to `oss/master`. If there are conflicts, resolve them,
    push to a `oss` branch, wait for CI, and then merge into `oss/master`.
    - If you make any additional changes, you are responsible for syncing these into `ent`.

### Build notes

**IMPORTANT:** Commands that build enterprise targets (e.g., `go build`, `yarn`,
`enterprise/dev/go-install.sh`) should always be run with the `enterprise` directory as the current
working directory. Otherwise, build tools like `yarn` and `go` may try to update the root
`package.json` and `go.mod` files as a side effect, instead of updating `enterprise/package.json`
and `enterprise/go.mod`.

The OSS web app is `yarn link`ed into `enterprise/node_modules`. It will run both the build of the
enterprise webapp as well as the part of the build for the OSS repo that generates the distributed
files for the npm package.

## Fallback syncing

Following the above instructions should prevent most long-term divergence between `oss` and
`ent`. Divergence between the two can easily be tested by running `dev/git-diff-no-enterprise.sh oss/master ent/master`.

If `oss` and `ent` diverge too severely to the point where it becomes onerous to cherry-pick
specific commits between the two, syncing can be accomplished by merging `oss/master` into
`ent/master`:

```
git fetch oss
git fetch ent
git checkout ent/master -b ent-master
git merge oss/master --no-commit -X theirs  # this allows us to view an accurate history of oss code in the enterprise repo
git checkout oss/master -- .                # overwrite all oss code in the enterprise repo with the state from oss repo
git reset enterprise && git checkout enterprise
git commit -m"Sync oss to ent $(date '+%Y-%m-%d')"
git push ent HEAD:master
```

Note that this means that `oss` is the source of truth for all code outside the `enterprise`
directory.
