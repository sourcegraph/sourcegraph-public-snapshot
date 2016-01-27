[![GoDoc](https://godoc.org/github.com/shazow/go-git?status.svg)](https://godoc.org/github.com/shazow/go-git)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/shazow/go-git/master/LICENSE)
[![Build Status](https://travis-ci.org/shazow/go-git.svg?branch=master)](https://travis-ci.org/shazow/go-git)

go-git
======

Pure-Go implementation of Git as a library.


## Origins

This is a fork of [gogits/git](https://github.com/gogits/git) (which was originally a fork of [speedata/gogit](https://github.com/speedata/gogit)).
Unfortunately, gogits/git is [no longer maintained and is not accepting PRs](https://github.com/gogits/git/issues/13#issuecomment-152784720), so this is a new extension of that work.

shazow/go-git is *not* backwards-compatible with gogits/git.


## Goals

### Short term

go-git's API should not be considered stable until these goals have been
achieved.

* [x] Remove features that depend on shelling out to git.
* [ ] Remove redundant features (in progress).
* [ ] Add key features to support implementing sourcegraph/go-vcs's Repository interface.
  (See [shazow/go-vcs#1](https://github.com/shazow/go-vcs/pull/1))
  - [x] Walk tree filesystem
  - [x] Walk commit ancestry
  - [ ] Diff
  - [ ] Blame
  - [ ] Subrepositories
  - [ ] Search


### Long term

* [ ] Add a storage driver interface for repositories with virtual/in-memory
      repository support.
* [ ] Improve test coverage.
* [ ] Improve query performance with caching and bitmap indexes.
* [ ] Rework use of locks (probably get rid of them, require consumer to manage
      repository locking).


## Sponsors

Work on this fork is sponsored by [Sourcegraph](https://sourcegraph.com/).


## License

* *shazow/go-git* changes are released under Apache v2.
* *gogits/git* (major rewrite of original MIT code) is licensed under Apache v2.
* *speedata/gogit* is licensed under MIT.
