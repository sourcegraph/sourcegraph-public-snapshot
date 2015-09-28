This file documents the style used in Sourcegraph's code.

For all things not covered in this document, defer to [Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments) and [Effective Go](http://golang.org/doc/effective_go.html).

### Panics
  Panics are used for code pathes that should never be reached.

### Options
  In the general case, when a pointer to an "options" struct is an argument to a function (such as `Get(build BuildSpec, opt *BuildGetOptions) (*Build, Response, error)`, that pointer may be `nil`. When the pointer is nil, the function does its default behavior. If the options struct should not be nil, either make the argument the value instead of a pointer or document it.

### Pull Requests
  When you have a PR that adds one new constant or other kind of declaration in a list that gofmt will automatically reindent, and if it changes the indentation, just add it in another block (separated with newlines). That way code reviewers don't have to wonder if you changed anything else in the block, and other people who might be working on the same code don't have to worry about lots of merge conflicts. Then as a separate commit on master, once the PR has been merged, you can delete the extraneous newlines and add it in its proper place.
