# Repository update frequency

By default, Sourcegraph polls code hosts to keep repository contents up to date, effectively running `git pull` periodically. You can also configure Sourcegraph to use [repository webhooks](webhooks.md), but this is usually not necessary.

The frequency at which Sourcegraph polls the code host for updates is determined by a smart hueristic based on past commit frequency in the repository. For example, if a repositories last commit was 8 hours ago, then the next update will be scheduled 4 hours from now. If there are still no new commits, then the next update will be scheduled 6 hours from then.

Repositories will never be updated more frequently than 45 seconds, and no less frequently than every 8 hours.

After Sourcegraph has polled the code host for Git updates, the search index will automatically update a short while after (usually taking just a few minutes per repository).

## Limiting repository updates

If you wish to control how frequently repositories are discovered or how frequently Sourcegraph polls your code host for updates, tuning parameters are available in the site configuration:

- [repoListUpdateInterval](../config/site_config.md#repoListUpdateInterval) controls how frequently we check the code host _for new repositories_ in minutes. Defaults to 1 minute.
- [gitMaxConcurrentClones](../config/site_config.md#gitMaxConcurrentClones) controls the maximum number of _concurrent_ cloning / pulling operations that Sourcegraph will perform.

You may also choose to disable automatic Git updates entirely and instead [configure repository webhooks](webhooks.md).
