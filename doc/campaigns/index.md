# Campaigns

<style>

.markdown-body h2 {
  margin-top: 2em;
}

.markdown-body ul {
  list-style:none;
  padding-left: 1em;
}

.markdown-body ul li {
  margin: 0.5em 0;
}

.markdown-body ul li:before {
  content: '';
  display: inline-block;
  height: 1.2em;
  width: 1em;
  background-size: contain;
  background-repeat: no-repeat;
  background-image: url(campaigns/file-icon.svg);
  margin-right: 0.5em;
  margin-bottom: -0.29em;
}

body.theme-dark .markdown-body ul li:before {
  filter: invert(50%);
}

</style>

<p class="subtitle">Make large-scale code changes across many repositories and code hosts</p>

<p class="lead">
Create a campaign by specifying a search query to get a list of repositories and a script to run in each. The campaign then lets you create pull requests on all affected repositories and tracks their progress until they're all merged. You can preview the changes and update them at any time.
</p>

<div class="cta-group">
<a class="btn btn-primary" href="quickstart">★ Quickstart</a>
<a class="btn" href="explanations/introduction_to_campaigns">Introduction to campaigns</a>
<a class="btn" href="references/requirements">Requirements</a>
</div>

> NOTE: This documentation describes the campaign functionality shipped in Sourcegraph 3.19 and src-cli 3.18, and later versions of both. [Click here](https://docs.sourcegraph.com/@3.18/user/campaigns) to read the documentation for campaigns in older versions of Sourcegraph and src-cli.
>
>
> We highly recommend using the latest versions of Sourcegraph and src-cli with campaigns, since we're steadily shipping new features and improvements.

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
- [Permissions in campaigns](explanations/permissions_in_campaigns.md)
- [Campaigns design](explanations/campaigns_design.md)
- [How `src` executes a campaign spec](explanations/how_src_executes_a_campaign_spec.md)

## How-tos

- [Creating a campaign](how-tos/creating_a_campaign.md)
- [Publishing changesets to the code host](how-tos/publishing_changesets.md)
- [Updating a campaign](how-tos/updating_a_campaign.md)
- [Viewing campaigns](how-tos/viewing_campaigns.md)
- [Tracking existing changesets](how-tos/tracking_existing_changesets.md)
- [Closing or deleting a campaign](how-tos/closing_or_deleting_a_campaign.md)
- [Site admin configuration for campaigns](how-tos/site_admin_configuration.md)
- [Configuring user credentials for campaigns](how-tos/configuring_user_credentials.md)
- <span class="badge badge-experimental">Experimental</span> [Creating multiple changesets in large repositories](how-tos/creating_multiple_changesets_in_large_repositories.md)

## Tutorials

- [Refactoring Go code using Comby](tutorials/refactor_go_comby.md)
- [Updating Go import statements using Comby](tutorials/updating_go_import_statements.md)
- [Update base images in Dockerfiles](tutorials/update_base_images_in_dockerfiles.md)
- [Search and replace specific terms](tutorials/search_and_replace_specific_terms.md)

## References

- [Requirements](references/requirements.md)
- [Campaign spec YAML reference](references/campaign_spec_yaml_reference.md)
- <span class="badge badge-experimental">Experimental</span> [Campaign spec templating](references/campaign_spec_templating.md)
- [Troubleshooting](references/troubleshooting.md)
