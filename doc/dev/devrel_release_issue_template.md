<!--
This template is used for tracking DevRel activities for our monthly major/minor release of Sourcegraph.
-->

# MAJOR.MINOR Release: DevRel tasks

<!-- 
  Once created, link the blog post pull request, email, and tweet for this release.
-->

- [ ] [Release blog post](#)
- [ ] [Release email](#)
- [ ] [Release tweet](#)

## At the start of the month (YYYY-MM-01)

- [ ] Create draft blog post Google doc in [Sourcegraph shared](https://drive.google.com/drive/u/0/folders/0B3lEU2lM-l9gUk5sNmRSMVFHVFU)
  - [ ] View each [team's release deliverables](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+milestone%3A{MAJOR}.{MINOR}+label%3Aroadmap) to generate the outline of the blog post
  - [ ] Share link to blog post doc in #progress Slack channel, asking Team leads to review

- [ ] Create draft tweet
- [ ] Create draft email
- [ ] Link to the blog post, email and tweet <!-- empty links are the top of the issue template>

### Create calendar events

Add events to the shared Release Schedule calendar in Google and invite team@sourcegraph.com:

- [ ] Blog post draft ready for review (5 days before release)
- [ ] Final blog post changes/suggestions must be submitted (3 days before the release)
- [ ] Blog post / tweets go live (the specific date we plan to do this, often not the 20th when it is a Friday).

## In the first week

- [ ] Confirm with each team that the planned deliverables are still on track to be announced in the blog post.
- [ ] Fill out each section based on the deliverables to first draft quality
- [ ] Think about what media (e.g. screenshot, screencast) will accompany the section content

## In the second week

- [ ] Remove `motd` previous release promotion from [Sourcegraph.com global settings](https://sourcegraph.com/site-admin/global-settings)
- [ ] Post the blog post draft to #dev-announce: `The blog post draft for <VERSION> is ready, please review your parts are accurate and provide feedback by <DATE>: <link>
- [ ] Tweet written
- [ ] Email written
- [ ] Write a warning at the top of the Google doc: `The blog post has been finalized and moved to Markdown, further changes here will not be reflected. Contact @ryan-blunden for suggestions.`
- [ ] Export blog post from Google docs to Markdown and create a new branch and draft pull request in [sourcegraph/about](https://github.com/sourcegraph/about/), using the [release blog post template](https://github.com/sourcegraph/about/blob/master/RELEASE_BLOG_POST_TEMPLATE.md)
- [ ] Send blog post, and email to [copy editor](https://docs.google.com/spreadsheets/d/1UUSSWrS8aKsLEg7M3Qdzw9s0GLJCI1eCrSJI06Qofb0/edit#gid=0_)

## 5 working days before release

- [ ] Start producing screenshots, diagrams, and screencasts for each blog post section

## 3 working days before release

- [ ] Check for any last minute release changes that affect the blog post, e.g. last minute feature removal
- [ ] Finalize screenshots, diagrams, and screencasts
- [ ] Blog post approved
- [ ] Email approved
- [ ] Tweet approved

## Day of release

- [ ] Publish blog post once the final release is cut, and docs version change is deployed
- [ ] Publish tweet:
  - [ ] Pin new release tweet
  - [ ] Confirm with Product if tweet will be promoted
- [ ] Send email in HubSpot
- [ ] Create new `notice` in [Sourcegraph.com global settings](https://sourcegraph.com/site-admin/global-settings)
   ```
   "notices": [
     {
       "message": "Sourcegraph {VERSION} is now available! Check out the [{VERSION} release blog post](https://about.sourcegraph.com/blog/sourcegraph-{VERSION}) for more details.",
       "location": "top",
       "dismissible": true
     }
   ]
   ```
- [ ] Use tweet or email content for post on the [LinkedIn Sourcegraph company page](https://www.linkedin.com/company/sourcegraph/)
- [ ] Put notification in #dev-rel Slack channel with links to blog post and tweet
