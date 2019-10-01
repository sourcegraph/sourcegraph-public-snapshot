# Sourcegraph roadmap

We want Sourcegraph to be:

- **For developers:** the best way to answer questions and get unblocked while writing, reviewing, or reading code.
- **For organizations** (engineering leaders and internal tools teams): the infrastructure for developer tools and data.

This roadmap is an overview of what’s coming in the next 3-6 months. The project details for how we get there are in our [Project Roadmap Google Doc](https://docs.google.com/document/d/1cBsE9801DcBF9chZyMnxRdolqM_1c2pPyGQz15QAvYI/edit?usp=sharing). Our high-level vision is outlined in the [Sourcegraph master plan](https://about.sourcegraph.com/plan).

We ship a release on the [20th day of each month](http://localhost:5080/dev/releases#releases-are-monthly).


Want to help us achieve these goals? [We're hiring!](https://github.com/sourcegraph/careers/blob/master/job-descriptions/software-engineer.md)

## Overview

We’re continually improving Sourcegraph’s core features for developers:

- Code search and navigation (with code intelligence)
- Integration into code review
- Automation of large-scale code changes

The current product focus is on the usability, scalability, and virality of the product. We strive to make Sourcegraph intuitive to use for anyone looking for answers about their code, and easy for admins to configure and maintain. We expect Sourcegraph to perform reliably at scale for our largest customers. We intend to create viral features that organically spread value throughout an organization.

## In 3-6 months, Sourcegraph will be

### User-friendly

Sourcegraph is intuitive to use for a wide range of roles, from developers to PMs, engineering managers, data analysts, and more. It adds value to this wide range of roles by making more information about code available, such as language statistics and dependency graphs. [Search has an improved UI](https://docs.google.com/document/d/1Vo7HlwO_HgrK8O-VEIZ9wHuSyHdEA0zk9qucNCoF0jg/edit?usp=sharing) that makes it more accessible, easier to use, and faster to drill down on what you're looking for.

Improving upon on our out-of-the-box code intelligence, Sourcegraph provides fast and precise code intelligence in the browser witha a quick and easy setup to your repository.

### Scalable, reliable, and secure

Sourcegraph performs at the scale of our largest customers under a wide variety of configurations and deployment infrastructure. [Code search is fast at large scale](https://docs.google.com/document/d/18w8T_KzYxQye8wg1g01QpMOX4_ERTtbOxMBRYaOEkmk/edit?usp=sharing) (~80k repositories), and the API can support enterprise level usage. Sourcegraph enforces repository permissions using ACLs from Bitbucket Server, GitHub, and GitLab.

Admins understand the health of their instances and are alerted proactively if things fail. It is clear what repositories are configured, and what their syncing status is.

### Viral

You can use Sourcegraph to [automate large-scale code changes](https://about.sourcegraph.com/product/automation) to remove legacy code, fix critical security issues, and pay down tech debt. Developers can collaborate on campaigns across thousands of repositories, and share, notify, and alert accordingly to get these changes pushed through.

<!--

Prior art:

https://about.gitlab.com/direction
https://docs.microsoft.com/en-us/visualstudio/productinfo/vs-roadmap
https://github.com/Microsoft/vscode/wiki/Roadmap

-->
