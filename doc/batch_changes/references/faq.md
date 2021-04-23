
This is a compilation of some common questions about Batch Changes.

### What happens if my batch change creation breaks down at 900 changesets out of 1,000? Do I have to re-run it all again?
Batch Changes' default behavior is to stop if creating the diff on a repo errors. You can choose to ignore errors instead by adding the [`-skip-errors`](../../cli/references/batch/preview.md) flag to the `src batch preview` command.

### Can we close a batch change and still leave the changesets open?
Yes. There is a confirmation page that shows you all the actions that will occur on the various changesets in the batch change after you close it. Open changesets will be marked 'Kept open', which means that batch change won't alter them. See [closing a batch change](../how-tos/closing_or_deleting_a_batch_change.md#closing-a-batch-change).

### How scalable are Batch Changes? How many changesets can I create?
Batch Changes can create tens of thousands of changesets. This is something we run testing on internally.
Known limitations:

- Since diffs are created locally by running a docker container, performance depends on the capacity of your machine. See [How `src` executes a batch spec](../explanations/how_src_executes_a_batch_spec.md).
- Manipulating (commenting, notifying users, etc) changesets at that scale can be clumsy. This is a major area of work for future releases.

### My batch change does not open changesets on all the repositories it should. Why?
- Do you have enough permissions? Batch Changes will error on the repositories you don’t have access to. See [Repository permissions for Batch Changes](../explanations/permissions_in_batch_changes.md).
- Does your `repositoriesMatchingQuery` contain all the necessary flags? If you copied the query from the sourcegraph UI, note that some flags are represented as buttons (case sensitivity, regex, structural search), and do not appear in the query unless the experimental [`copyQueryButton`](https://github.com/sourcegraph/sourcegraph/pull/18317) feature toggle is enabled.

### Can I create tickets or issues along with Batch Changes?
Batch Changes does not support a declarative syntax for issues or tickets.
However, [steps](../references/batch_spec_yaml_reference.md#steps-run) can be used to run any container. Some users have built scripts to create tickets at each apply:

- [Jira tickets](https://github.com/sourcegraph/campaign-examples/tree/master/jira-tickets)
- [GitHub issues](https://github.com/sourcegraph/batch-change-examples/tree/main/github-issues)

### What happens to the preview page if the batch spec is not applied?
Unapplied batch specs are removed from the database after 7 days.

### Can I pull containers from private container registries in a batch change?
When [executing a batch spec](../explanations/how_src_executes_a_batch_spec.md), `src` will pull from the current container registry. If you are logged into a private container registry, it will pull from it.

### What tool can I use for changing/refactoring `<programming-language>`?

Batch Changes supports any tool that can run in a container and changes file contents on disk. You can use the tool/script that works for your stack or build your own, but here is a list of [examples](https://github.com/sourcegraph/batch-change-examples) to get started.
Common language agnostic starting points:

- `sed`, `yq`, `awk` are common utilities for changing text
- [comby](https://comby.dev/docs/overview) is a language-aware structural code search and replace tool. It can match expressions and function blocks, and is great for more complex changes.
