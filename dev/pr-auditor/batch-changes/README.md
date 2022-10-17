# pr-auditor batch changes

## Rollouts

[`pr-auditor-rollout.yml`](./pr-auditor-rollout.yml) describes a batch change that can be used to roll out `pr-auditor` to repositories that do not have it set up yet.

To use it:

1. [Create a batch change on `/batch-changes/create`](https://k8s.sgdev.org/batch-changes/create)
2. Paste [`pr-auditor-rollout.yml`](./pr-auditor-rollout.yml) into the spec input and make the appropriate changes to the template (for example, adjust the `on:` parameters to target the desired repositories)
3. Preview and apply the changes
4. Selectively publish the appropriate changesets for review and merge

## Updates

The [`pr-auditor-patch.yml`](./pr-auditor-patch.yml) spec describes a batch change that can be used to roll out `pr-auditor` workflow updates to repositories that already have it set up by synchronizing them with the source-of-truth script in the batch change spec.
