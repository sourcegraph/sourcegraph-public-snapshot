# Using Sourcegraph with a monorepo

Sourcegraph can be used on Git [monorepos](https://trunkbaseddevelopment.com/monorepos/) and provides functionality that developers may be missing:

- Indexed code search across projects in the monorepo.
- Code intelligence across projects in the monorepo.

The tools on a developer's local machine often can't cope with the scale of a monorepo. For example, a developer often cannot load the whole monorepo into their IDE, and therefore has to do without code exploration functionality useful when doing cross-project work. Sourcegraph brings this functionality back to developers, enabling them to work efficiently at the day-to-day task of understanding code.

## Scale and performance

Sourcegraph is designed to scale to large monorepos, even when local Git operations and other code search tools become slow.

Sourcegraph uses the standard `git` binary to interact with repositories, but structures its use of `git` to ensure that user requests remain fast. Commands like `git rev-parse` and `git show` are called when rendering a page for a user. Git operations that scale with the total number of files are left to background indexing jobs. Operations that scale with the total number of commits or refs/tags may be in the user path, but these operations use pagination to remain fast.

Sourcegraph's code search index scales horizontally with the number of files being indexed for search. Multiple shards may be allocated for one repository, and the index is agnostic to whether the code exists in one massive repository or many smaller ones. Sourcegraph has been used to index both large monorepos and tens of thousands of smaller repositories.

<!-- TODO code intelligence LSIF -->

## Custom git binaries

Sourcegraph clones code from your code host via the usual `git clone` or `git fetch` commands. Some organisations use custom `git` binaries or commands to speed up these operations. Sourcegraph supports using alternative git binaries to allow cloning. This can be done by inheriting from the `gitserver` docker image and installing the custom `git` onto the `$PATH`.

<!-- More TODOs

- [ ] Usage of Sourcegraph with git submodules or sub-trees (how do these appear in the UI/in searches?)
- [ ] git-meta
- [ ] Massive numbers of branches
- [ ] UI tweaks (e.g., "can we customize the search scopes to show submodules or branches instead of repos?")

-->
