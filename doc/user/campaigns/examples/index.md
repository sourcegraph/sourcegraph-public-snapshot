# Example campaigns

> NOTE: This documentation describes the current work-in-progress version of campaigns. [Click here](https://docs.sourcegraph.com/@3.18/user/campaigns) to read the documentation for campaigns in Sourcegraph 3.18.

<!-- TODO(sqs): update for new campaigns flow -->

The following examples demonstrate various types of campaigns for different languages using both commands and Docker images. They also provide commentary on considerations such as adjusting the duration (`-timeout`) for actions that exceed the 15 minute default limit.

* [Using ESLint to automatically migrate to a new TypeScript version](./eslint_typescript_version.md)
* [Adding a GitHub action to upload LSIF data to Sourcegraph](./lsif_action.md)
* [Refactoring Go code using Comby](./refactor_go_comby.md)
