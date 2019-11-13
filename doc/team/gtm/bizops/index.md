# Analytics

This page describes Sourcegraph's analytics function, our data sources, and how to use our data tools.

## How to submit a data request

[Looker](https://sourcegraph.looker.com/) is a self-service tool so we encourage everyone to try finding answers within the tool. [Get started here](#using-looker) if you're not familiar with Looker, and the team in the #analytics Slack channel are more than happy to answer any questions regarding Looker. 

**Projects:** Add an issue to the [analytics board](https://github.com/orgs/sourcegraph/projects/63) in GitHub. This is used as the analytics project board and is triaged everyday. Please think through and include the following so we can effectively prioritize your request.
- What is the deliverable going to be used for? Why do you need it? This helps the team prioritize requests. 
- What do you want the deliverable to look like (in as much detail as possible)? For example, if you want a chart it would be extremely helpful to draw the chart on paper and attach it.
- When do you need the deliverable by? 
- Is this request a nice-to-have or a necessity?

**Small asks and questions:** Post in the #analytics channel in Slack. 

## Data sources

Here are the following sources we collect data from:

* [Looker](#using-looker): Business intelligence/data visualization tool
* Google Analytics: Website analytics for Sourcegraph marketing and docs pages (not Sourcegraph.com)
* Google Cloud Platform: BigQuery is our data warehouse and the database Looker runs on top of
* Google Sheets: There are a [number of spreadsheets](https://drive.google.com/drive/folders/1vOyhFO90FjHe-bwnHOZeljHLuhXL2BAv)that Looker queries (by way of BigQuery).
* HubSpot: Marketing automation and CRM
* Apollo: Email marketing automation
* Sourcegraph.com Site-admin pages: customer subscriptions and license keys
* [Pings](https://docs.sourcegraph.com/admin/pings) from self-hosted Sourcegraph instances containing anonymous and aggregated information
* [Custom tool to track events](https://github.com/sourcegraph/sourcegraph/issues/5486) on Sourcegraph.com instance

## Using Looker

Coming soon
