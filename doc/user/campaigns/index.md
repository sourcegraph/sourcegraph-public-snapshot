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

.getting-started-box {
  flex: 1;
  margin: 0.5em;
  padding: 1rem 1.25rem;
  border-radius: 4px;
  border: 1px solid var(--sidebar-nav-active-bg);
}
.getting-started-box a {
  color: var(--text-color);
}

.getting-started-box a:hover {
  color: var(--link-color);
}

.getting-started-box a span {
  color: var(--link-color);
  font-weight: bold;
}

</style>

<p class="subtitle">Make large-scale code changes across many repositories and code hosts</p>

<p class="lead">
Create a campaign by specifying a search query to get a list of repositories and a script to run in each. The campaign then lets you create pull requests on all affected repositories and tracks their progress until they're all merged. You can preview the changes and update them at any time.
</p>

<div class="cta-group">
<a class="btn btn-primary" href="quickstart">★ Quickstart</a> <a class="btn" href="introduction_to_campaigns">Introduction to campaigns</a>
</div>

> NOTE: This documentation describes the campaign functionality shipped in Sourcegraph 3.19 and src-cli 3.18. [Click here](https://docs.sourcegraph.com/@3.18/user/campaigns) to read the documentation for campaigns in older versions of Sourcegraph and src-cli.

## Getting started

<div class="getting-started">
  <div class="getting-started-box">
  <a href="quickstart">
  <span>New to campaigns?</span></br>Run through the <b>quickstart guide</b> and create a campaign in less than 10 minutes.
  </a>
  </div>
  <div class="getting-started-box">
  <a href="https://www.youtube.com/watch?v=EfKwKFzOs3E">
  <span>Demo video</span></br>Watch the campaigns demo video to see what campaigns are capable of.
  </a>
  </div>
  <div class="getting-started-box">
  <a href="introduction_to_campaigns">
  <span>Introduction to campaigns</span></br>Find out what campaigns are and what they can, learn key concepts and see what others use them for.
  </a>
  </div>
</div>

## How-tos

- [Viewing campaigns](how-tos/viewing_campaigns.md)
- [Creating a campaign](how-tos/creating_a_campaign.md)
- [Publishing changesets to the code host](how-tos/publishing_changesets.md)
- [Tracking existing changesets](how-tos/tracking_existing_changesets.md)
- [Closing or deleting a campaign](how-tos/closing_or_deleting_a_campaign.md)
- [Code host and repository permissions in campaigns](how-tos/code_host_repository_permissions.md)
- [Managing access to campaigns](how-tos/managing_access.md)
- [Site admin configuration for campaigns](how-tos/site_admin_configuration.md)

## Tutorials

- [Refactoring Go code using Comby](tutorials/refactor_go_comby.md)
- [Updating Go import statements using Comby](tutorials/updating_go_import_statements.md)

## References

- [Campaign spec YAML reference](campaign_spec_yaml_reference.md)

## Known issues

- Campaigns currently support **GitHub**, **GitLab** and **Bitbucket Server** repositories. If you're interested in using campaigns on other code hosts, [let us know](https://about.sourcegraph.com/contact).
- It is not yet possible for a campaign to create multiple changesets in a single repository (e.g., to make changes to multiple subtrees in a monorepo).
- Forking a repository and creating a pull request on the fork is not yet supported. Because of this limitation, you need write access to each repository that your campaign will change (in order to push a branch to it).
- Campaign steps are run locally (in the [Sourcegraph CLI](https://github.com/sourcegraph/src-cli)). Sourcegraph does not yet support executing campaign steps on the server. For this reason, the APIs for creating and updating a campaign require you to upload all of the changeset specs (which are produced by executing the campaign spec locally). {#server-execution}
- It is not yet possible for multiple users to edit the same campaign that was created under an organization.
- It is not yet possible to reuse a branch in a repository across multiple campaigns.
