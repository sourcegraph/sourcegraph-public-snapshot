This file documents the style used in Sourcegraph's code and product.

For all things not covered in this document, defer to
[Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments)
and [Effective Go](http://golang.org/doc/effective_go.html).

# English

The same standards apply to documentation, commit messages, code
comments, and user interface text.

* Phrases in headers and titles are capitalized like sentences ("Add
  repository"), not headlines ("Add Repository").
* Use descriptive link text, such as "Need to
  [view your repositories](https://sourcegraph.com/sourcegraph/sourcegraph@master/-/blob/docs/style.md#)?" Don't use "here" as link text, as in
  "Need to view your repositories? [Click here.](https://sourcegraph.com/sourcegraph/sourcegraph@master/-/blob/docs/style.md#)"

## Terms

* In UI text and documentation, always use "repository" instead of
  "repo."

## Names

Use the standard rendering of names in prose. For example:

* [PostgreSQL](http://www.postgresql.org/about/), not Postgres,
  postgres, PgSQL, Postgresql, PostGres, etc. (Using something like
  `package pgsql` in code is fine, but be consistent: don't name one
  package `postgres` and the other `pgsql`.)
* Sourcegraph, not sourcegraph or SourceGraph.
* Go, not Golang.
* GitHub, not Github.
* OS X, not OSX.
* gRPC, not grpc or GRPC.
* Bitbucket, not BitBucket.
* JIRA, not Jira.
* Docker, not docker. (If you are referring to the `docker` CLI tool,
  then that is not in prose.)

When in doubt, Google it.

# Code

## Panics

Panics are used for code pathes that should never be reached.

## Options

In the general case, when a pointer to an "options" struct is an argument
to a function (such as `Get(build BuildSpec, opt *BuildGetOptions) (*Build, Response, error)`,
that pointer may be `nil`. When the pointer is nil, the function does its default behavior.
If the options struct should not be nil, either make the argument the value instead of a
pointer or document it.

## Changesets

When you have a CS that adds one new constant or other kind of declaration in a list that
gofmt will automatically reindent, and if it changes the indentation, just add it in another
block (separated with newlines). That way code reviewers don't have to wonder if you changed
anything else in the block, and other people who might be working on the same code don't have
to worry about lots of merge conflicts. Then as a separate commit on master, once the CS has
been merged, you can delete the extraneous newlines and add it in its proper place.
