# Documentation structure

> NOTE: Adapted from [GitLab documentation guidelines](https://gitlab.com/gitlab-org/gitlab-ee/raw/master/doc/development/documentation/structure.md).

For consistency throughout the documentation, it's important to maintain the same
structure among the docs.

Before getting started, read through the following docs:

- [Contributing to Sourcegraph documentation](index.md#contributing-to-docs)
- [Documentation directory structure](index.md#documentation-directory-structure)
- [Documentation style guidelines](style_guide.md)
- Example Documentation: [How to setup a separate website maintained by Sourcegraph](separate_website.md)

## Documentation blurb

Every document should include the following content in the following sequence:

- **Feature name**: defines an intuitive name for the feature that clearly
  states what it is and is consistent with any relevant UI text.
- **Feature overview** and description: describe what it is, what it does, and in what context it should be used.
- **Use cases**: describes real use case scenarios for that feature.
- **Requirements**: describes what software and/or configuration is required to be able to
  use the feature and, if applicable, prerequisite knowledge for being able to follow/implement the tutorial.
  For example, familiarity with Kubernetes, an account on a third-party service, dependencies installed, etc.
  Link each one to its most relevant resource; i.e., where the reader can go to begin to fulfill that requirement.
  (Another doc page, a third party application's site, etc.)
- **Instructions**: clearly describes the steps to use the feature, leaving no gaps.
- **Troubleshooting** guide (recommended but not required): if you know beforehand what issues
  one might have when setting it up, or when something is changed, or on upgrading, it's
  important to describe those too. Think of things that may go wrong and include them in the
  docs. This is important to minimize requests for support, and to avoid doc comments with
  questions that you know someone might ask. Answering them beforehand only makes your
  document better and more approachable.

For additional details, see the subsections below, as well as the [Documentation template for new docs](#Documentation-template-for-new-docs).

### Feature overview and use cases

Every major feature should have, at the beginning of the document, two main sections: **Overview** and **Use cases**.

**Overview**: as the name suggests, the goal here is to provide an overview of the feature. Describe what is it, what it does, why it is important/cool/nice-to-have, what problem it solves, and what you can do with this feature that you couldn't do before.

**Use cases**: provide at least two, ideally three, use cases for every major feature. You should answer this question: what can you do with this feature/change? Use cases are examples of how this feature or change can be used in real life.

Note that if you don't have anything to add between the doc title (`<h1>`) and the header `## Overview`, you can omit the header, but keep the content of the overview there.

> The overview and use cases are required for **every** Enterprise feature, and for every **major** feature present in OSS.

### Discoverability

Your new document will be discoverable by the user only if:

- Crosslinked from the higher-level index (e.g., Google Cloud Docker deployment docs should be linked from Docker deployment docs)
  - When referencing other Sourcegraph products and features, link to their respective docs; when referencing third-party products or technologies, link to their external sites, documentation, and resources.
  - The headings are clear. E.g., "Google Cloud Deployment" is a bad heading, "Deploying Sourcegraph to Google Cloud using Docker" is much better. Think of the keywords that a user would search for, and use those keywords in the headings.

## Documentation template for new docs

To start a new document, respect the file tree and file name guidelines, as well as the style guidelines. Use the following template:

```md
---
description: "short document description." # Up to ~200 chars long. They will be displayed in Google search result previews.
---

# Feature name

> [Introduced](link) in Sourcegraph X.Y.

A short description for the feature (can be the same used in the frontmatter's
`description`).

## Overview

To write the feature overview, you should consider answering the following questions:

- What is it?
- Who is it for?
- What is the context in which it is used and are there any prerequisites/requirements?
- What can the user do with it? (Be sure to consider multiple audiences, like Sourcegraph site admins and regular users.)
- What are the benefits to using it over any alternatives?

## Use cases

Describe one to three use cases for that feature. Give real-life examples.

## Requirements

State any requirements, if any, for using the feature and/or following along with the tutorial.

The only assumption that is redundant and doesn't need to be mentioned is having an account on Sourcegraph.

## Instructions

("Instructions" is not necessarily the name of the heading)

- Write a step-by-step guide, with no gaps between the steps.
- Start with an h2 (`##`), break complex steps into small steps using
  subheadings h3 > h4 > h5 > h6. _Never skip the hierarchy level, such
  as h2 > h4_, as it will break the TOC and may affect the breadcrumbs.
- Use short and descriptive headings (up to ~50 chars). You can use one
  single heading `## How it works` for the instructions when the feature
  is simple and the document is short.
- Be clear, concise, and stick to the goal of the doc: explain how to
  use that feature.
- Use inclusive language and avoid jargons, as well as uncommon and
  fancy words. The docs should be clear and easy to understand.
- Write in the 3rd person (use "we", "you", "us", "one", instead of "I" or "me").
- Always provide internal and external reference links.
- Always link the doc from its higher-level index.

<!-- ## Troubleshooting

Add a troubleshooting guide when possible/applicable. -->
```

Notes:

- Apply the correct format for the [Sourcegraph version introducing the feature](style_guide.md#sourcegraph-versions-and-tiers)
