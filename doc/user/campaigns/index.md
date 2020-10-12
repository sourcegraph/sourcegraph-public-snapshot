# Campaigns

<style>

.subtitle {
  font-weight: 600;
  margin-top: -0.5em;
  font-size: 1.3rem;
  color: var(--text-muted);
}

.lead {
  font-size: 1.15rem;
}

.markdown-body h2 {
  margin-top: 2em;
}

.btn {
  display: inline-block;
  margin: 0;
  padding: 1rem 1.25rem;
  border-radius: 4px;
  text-decoration: none;
  -webkit-appearance: none;
  -webkit-font-smoothing: antialiased;
  border: 1px solid var(--sidebar-nav-active-bg);
}

.btn-primary {
  background-color: var(--sidebar-nav-active-bg);
}

.btn:hover {
  opacity: 0.85;
  text-decoration: none;
}

.cta-group {
  margin: 3em 0;
}

.getting-started {
  display: flex;
  align-items: stretch;
}

.getting-started .btn {
  flex: 1;
  margin: 0.5em;
  padding: 1rem 1.25rem;
  color: var(--text-color);
  border-radius: 4px;
  border: 1px solid var(--sidebar-nav-active-bg);
}

.getting-started .btn:hover {
  color: var(--link-color);
}

.getting-started .btn span {
  color: var(--link-color);
  font-weight: bold;
}

</style>

<p class="subtitle">Make large-scale code changes across many repositories and code hosts</p>

<p class="lead">
Create a campaign by specifying a search query to get a list of repositories and a script to run in each. The campaign then lets you create pull requests on all affected repositories and tracks their progress until they're all merged. You can preview the changes and update them at any time.
</p>

<div class="cta-group">
<a class="btn btn-primary" href="quickstart">★ Quickstart</a>
<a class="btn" href="explanations/introduction_to_campaigns">Introduction to campaigns</a>
</div>

> NOTE: This documentation describes the campaign functionality shipped in Sourcegraph 3.19 and src-cli 3.18. [Click here](https://docs.sourcegraph.com/@3.18/user/campaigns) to read the documentation for campaigns in older versions of Sourcegraph and src-cli.

## Getting started

<div class="getting-started">
  <a href="quickstart" class="btn" alt="Run through the Quickstart guide">
   <span>New to campaigns?</span>
   </br>
   Run through the <b>quickstart guide</b> and create a campaign in less than 10 minutes.
  </a>

  <a href="https://www.youtube.com/watch?v=EfKwKFzOs3E" class="btn" alt="Watch the campaigns demo video">
   <span>Demo video</span>
   </br>
   Watch the campaigns demo video to see what campaigns are capable of.
  </a>

  <a href="explanations/introduction_to_campaigns" class="btn" alt="Read the Introduction to campaigns">
   <span>Introduction to campaigns</span>
   </br>
   Find out what campaigns are, learn key concepts and see what others use them for.
  </a>
</div>

## Explanations

- [Introduction to campaigns](explanations/introduction_to_campaigns.md)
- [Managing access to campaigns](explanations/permissions_in_campaigns.md)

## How-tos

- [Creating a campaign](how-tos/creating_a_campaign.md)
- [Publishing changesets to the code host](how-tos/publishing_changesets.md)
- [Updating a campaign](how-tos/updating_a_campaign.md)
- [Viewing campaigns](how-tos/viewing_campaigns.md)
- [Tracking existing changesets](how-tos/tracking_existing_changesets.md)
- [Closing or deleting a campaign](how-tos/closing_or_deleting_a_campaign.md)
- [Site admin configuration for campaigns](how-tos/site_admin_configuration.md)

## Tutorials

- [Refactoring Go code using Comby](tutorials/refactor_go_comby.md)
- [Updating Go import statements using Comby](tutorials/updating_go_import_statements.md)

## References

- [Campaign spec YAML reference](campaign_spec_yaml_reference.md)
