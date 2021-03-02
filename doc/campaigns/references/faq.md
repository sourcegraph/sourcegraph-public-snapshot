
This is a compilation of some common questions about campaigns.

### What happens if my campaign creation breaks down at 900 changesets out of 1,000? Do I have to re-run it all again?
Campaigns’ default behavior is to stop if creating the diff on a repo errors. You can choose to ignore errors instead by adding the [`-skip-errors`](../../cli/references/campaigns/preview.md) flag to the `src campaigns preview` command.

### Can we close a campaign and still leave the changesets open?
Yes. There is a confirmation page that shows you all the actions that will occur on the various changesets in the campaign after you close it. Open changesets will be marked 'Kept open', which means that campaigns won't alter them. See [closing a campaign](../how-tos/closing_or_deleting_a_campaign.md#closing-a-campaign).

### How scalable are campaigns? How many changesets can I create?
Campaigns can create tens of thousands of changesets. This is something we run testing on internally.
Known limitations:

- Since diffs are created locally by running a docker container, performance depends on the capacity of your machine. See [How `src` executes a campaign spec](../explanations/how_src_executes_a_campaign_spec.md).
- Manipulating (commenting, notifying users, etc) changesets at that scale can be clumsy. This is a major area of work for future releases.

### My campaign does not open changesets on all the repositories it should. Why?

- Do you have enough permissions? Campaigns will error on the repositories you don’t have access to. See [Repository permissions for campaigns](../explanations/permissions_in_campaigns.md).

- Does your `repositoriesMatchingQuery` contain all the necessary flags? If you copied the query from the sourcegraph UI, note that some flags are represented as buttons (case sensitivity, regex, structural search), and do not appear in the query unless the experimental [`copyQueryButton`](https://github.com/sourcegraph/sourcegraph/pull/18317) feature toggle is enabled.

### Can I create tickets or issues along with campaigns?
Campaigns does not support a declarative syntax for issues or tickets.
However, [steps](../references/campaign_spec_yaml_reference.md#steps-run) can be used to run any container. Some users have built scripts ([example](https://github.com/sourcegraph/campaign-examples/tree/master/jira-tickets)) to create tickets at each apply.

### What happens to the preview page if the campaign spec is not applied?
Unapplied campaign specs are removed from the database after 7 days.
