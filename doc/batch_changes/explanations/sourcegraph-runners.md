# Sourcegraph runner

By default, Batch Changes uses a command line interface in your local environment to compute diffs and create changesets. This can be impractical for creating large batch changes that require a lot of processing, or if you don't want to use your local environment for batch changes. If you are on a developer tool team rolling out batch changes to your company, the requirement to create batch changes locally might make adoption more difficult for some of your users.

**Sourcegraph runner is an application that allows users to offload computing batch changes to dedicated infrastructure.**

- If you use Sourcegraph enterprise on-premise, you can install Sourcegraph runner on infrastructure you own or manage.
- If you use a managed Sourcegraph instance, TODO.
- Sourcegraph.com offers a batch change execution environment based on Sourcegraph runner. You don't have to do anything to use it.

Sourcegraph runner is [open source](TODO).

# Prerequisites

- You can install Sourcegraph runner in the following environments: TODO, TODO, TODO
- Sourcegraph runner requires Docker to be installed
- Sourcegraph runner requires requires Sourcegraph version xxx.


# Registering a runner

TODO

# Using runners

TODO

## Who has access to runners

TODO

## Debugging

TODO

# Administering and monitoring runners

TODO

# How do runners work

TODO

# Runner vs CLI workflow comparison

TODO
