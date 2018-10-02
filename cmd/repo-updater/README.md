# repo-updater architecture/dev notes

This document is here because repo-updater has a lot of historical quirks and architecture choices that are easy to miss on first inspection, and I would have liked it if this had existed when I first started reading the code. Some of what's documented here is historical stuff that needs to be refactored, but explaining where we are now makes it easier to see what's going on.

repo-updater evolved into what it is now, and if you've ever studied biology, this explains everything. We are currently in the process of refining and refactoring repo-updater; this started with work on scheduling, but there's a lot of loose ends and things that need to get documented, organized, and possibly reworked.

## Design history

Repo-updater has two primary tasks: Maintaining metadata about repositories for other components to look up, and making sure repositories get periodic fetches run against them to keep them close to in sync with upstream.

### Repo metadata and lookups

The primary function of repo-updater for repository metadata is to provide an abstraction layer and cache on top of the interfaces of various different code hosts. As of this writing, it has support for:

- AWS Code Commit
- Bitbucket Server
- GitHub
- GitLab
- Gitolite
- Phabricator
- repos.list information in config.json

The exposed API call `RepoLookup` is the primary interface to this. Everything else calls RepoLookup when it wants to find out about a repository. This function then tries each of the code hosts in a consistent order, looking for one which claims to be authoritative. The current order is:

- GitHub
- GitLab
- Bitbucket Server
- AWS Code Commit
- repos.list information in config.json
- Gitolite

Note that Phabricator isn't listed here. (As of this writing, I don't know why.)

The implementations of these mostly live in `repos/*.go`, but GitHub, GitLab, Bitbucket Server, and AWS Code Commit also have code in `internal/externalservice/[servicename]`.

In each of these cases, the code in `internal/externalservice` implements a cache using `sourcegraph/pkg/rcache`, which is supposed to be an application-level cache. There's also an HTTP cache used for the actual transport. In the specific case of github, API calls are actually proxied through `github-proxy`. The HTTP cache handles things like `Etag` and `If-Modified-Since` headers to reduce spamminess of API calls; GitHub, for instance, won't count an API call as consuming API rate limit if it returns a 304.

In general, the spamminess of these requests is a significant concern; I did a simple search and saw >10 requests in a row to github about the _same_ repository in a second or two. (Inference: The caching is unreliable or not working as intended.)

### Fetching repositories

The repository fetching work started out as a simple "every N minutes, run a fetch on every repo" process. This was not necessarily ideal, and scaled poorly. Meanwhile, requests from other components for updates were sent directly to `gitserver`, which did some debouncing (not fetching things that had been fetched in the last ten seconds), but this didn't prevent the periodic updates from being duplicative of other updates.

The migration process started moving everything else to ask repo-updater to do updates, but repo-updater just forwarded that to gitserver.

But now that everything goes through repo-updater, it can notice when those happen, and reschedule periodic updates, which are now scheduled per-repository. Furthermore, the intervals at which they're checked can vary based on other criteria; repositories which rarely see updates don't get checked as often as frequently-updated repositories.

This is still being worked on and revised. For historical reasons, the logic for the scheduler is all living in `repos/reposlist.go`, which is an error; that should be the source for handling the specifics of the `repos.list` property of site config. This should get addressed in a future iteration.

## Future directions

The scheduler should move out of the repos/ directory entirely.

The code doing the iterative lookups should probably be a loop, or something, with some registration of things. Same for the `[Host]SyncWorker` threads; we should have a standard interface to code hosts that they expose.

The caching code in `internal/externalservice/*` appears to be very similar, and should perhaps be factored out.

We need better metrics on the background fetching loop, and suitable controls to adjust its behavior.

This file needs to get updated when those things happen.
