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
