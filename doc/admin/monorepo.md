# Using Sourcegraph with a monorepo

Sourcegraph can be used on git [monorepos](https://trunkbaseddevelopment.com/monorepos/) and provides functionality that developers may be missing:

- Indexed code search across projects in the monorepo.
- Code intelligence across projects in the monorepo.

This functionality may be missing since the tools on a developers machines often can't cope with the scale of a monorepo. Normally a developer can't load the whole monorepo into your IDE losing functionality useful when doing cross-project work.

## Scale/performance concerns

Sourcegraph ships with the standard git binary to interact with repositories. Commands like `git rev-parse` and `git show` are called when rendering a page for a user. Operations which scale with the total number of files are left to background indexing jobs. Operations which scale with the total number of commits or refs/tags can be in the user path. These operations use pagination to avoid the scale concerns.

Our indexed search scales by the number of files so has no particular concerns due to monorepos. Sourcegraph has been used to index large monorepos as well as tens of thousands of smaller repositories.

<!-- TODO code intelligence LSIF -->

## Custom git binaries

Sourcegraph clones code from your code host via the usual `git clone` or `git fetch` commands. Some organisations use custom `git` binaries or commands to speed up these operations. Sourcegraph supports using alternative git binaries to allow cloning. This can be done by inheriting from the `gitserver` docker image and installing the custom `git` onto the `$PATH`.

<!-- More TODOs

- [ ] Usage of Sourcegraph with git submodules or sub-trees (how do these appear in the UI/in searches?)
- [ ] git-meta
- [ ] Massive numbers of branches
- [ ] UI tweaks (e.g., "can we customize the search scopes to show submodules or branches instead of repos?")

-->
