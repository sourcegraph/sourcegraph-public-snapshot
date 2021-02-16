### What happens if my campaign creation breaks down at 900 changeset out of 1,000? Do I have to re-run it all again?
Campaigns’ default behavior is to stop if creating diffs on a repo errors. You can choose to ignore errors instead by adding the `[skip-errors](../cli/references/campaigns/preview)` flag to the `src campaigns` command.

### Can we close a campaign and still leave the changesets open?
Yes! There is a confirmation page that shows you all the actions that will occur on the various changesets in the campaign after you close it.
Open changesets will be marked 'Kept open', which means that campaigns won't alter them.

### How scalable is Campaigns? How many changesets can I create?
Campaigns can create tens of thousands changesets. This is something we run testing on internally. Known limitations:
Since diffs are created locally by the src cli, performance depends on the capacity of your machine.
Manipulating (commenting, notifying users, etc) changesets at that scale can be clumsy. This is a major area of work for future releases

### Can I run campaigns in CI?
Yes. Some of our users have a repository with campaign specs

### My campaign does not open changesets on all the repositories it should. Why?
Do you have enough permissions? Campaigns will error on the repositories you don’t have access to. See [codehost permissions]().
Does your repositoriesMatchingQuery contain all the necessary flags? If you copied the query from the sourcegraph UI, note that some flags are represented as buttons (case sensitivity, regex, structural search).

### What happens If the user generating the changeset does not have access to all of the repositories?
TODO
