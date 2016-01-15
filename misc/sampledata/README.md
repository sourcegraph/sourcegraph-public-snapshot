# sampledata

This directory contains sample Git repositories that are added to new
Sourcegraph servers during onboarding.

## Working with the sample repositories

We want to keep these sample Git repositories tracked in the
Sourcegraph repository without the additional complexity of, e.g., Git
submodules.

To store these Git repositories inside the main Sourcegraph
repository, use a trick: rename the sample repositories' `.git`
directories to `.git-versioned` before you commit. When you need to
work with the sample repositories, temporarily rename them back to
`.git`.

When committing, use the following command so that the committer and author are Sourcegraph Demo:

```
GIT_AUTHOR_NAME="Sourcegraph Demo" GIT_AUTHOR_EMAIL="demo@sourcegraph.com" GIT_COMMITTER_NAME="Sourcegraph Demo" GIT_COMMITTER_EMAIL="demo@sourcegraph.com" git commit ...
```

Verify that the author and committer entries are "Sourcegraph Demo <demo@sourcegraph.com>" with `git log --format=full`.

Finally, run `git gc --aggressive --prune=all` after you commit.
