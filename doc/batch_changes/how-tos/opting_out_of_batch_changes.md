# Opting out of batch changes

> NOTE: This feature is requires [Sourcegraph CLI](https://github.com/sourcegraph/src-cli) 3.26.3 or later.

Repository owners that are not interested in batch change changesets can opt out
so that their repository will be skipped when a batch spec is executed.

1. Create a file called `.batchignore` in the repository you wish to be skipped.
2. `src batch [apply|preview]` will now skip that repository if it's yielded by the `on` part of the batch spec.
3. Use the `-force-override-ignore` flag to override that behaviour and not skip any ignored repositories.
