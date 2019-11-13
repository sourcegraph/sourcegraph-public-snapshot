# Analytics

## How to submit a data request

**Large projects:** Add an issue to the left column of this project in GitHub. This is triaged everyday. Please include the following information:
* What is the deliverable going to be used for? Why do you need it? This helps the team prioritize requests. 
* What do you want the deliverable to look like (in as much detail as possible)? For example, If you want a chart you could draw the chart on paper and attach it.
* When do you need the deliverable by? 
* Is this request a nice-to-have or a necessity?

**Small asks and questions:** post in the #analytics channel in Slack. 

## Go-to-market tools

Here are the following tools we use:

* [Looker](#using-looker): Business intelligence/data visualization tool
* Google Analytics: Website analytics
* Google Cloud Platform: BigQuery as our data warehouse and the database our Looker runs on top of
* Google Sheets: There are a number of spreadsheets that Looker queries (by way of BigQuery). You can find them in the shared Google Drive under ‘Analytics’ -> ‘ETL’
* [HubSpot](#marketing): Marketing automation and CRM
* Apollo: Email marketing automation
* Sourcegraph Admin: Customers, subscriptions, user surveys, usage stats, etc…
* Custom event tool to track event

## Data sources

###Product

Private instances (on-prem)

We receive pings every 30 minutes than [contains this data](https://docs.sourcegraph.com/admin/pings). 

#### Sourcegraph.com

"All" frontend user actions are logged inside of the Sourcegraph instance. This includes everything from "user viewed a page" to "user clicked a button" to "user hovered over a symbol". You can look at all calls to eventLogger.log here. 

All of these events are stored in a backend database containing:
* name: the name of the event
* argument (or whatever): the string argument
* url: the URL on the page when the event was logged
* userID: the user that took the action (null if not signed in)
* anonymousUserID: ID assigned to the user regardless of whether they are signed in or anonymous (through a UUID stored in a cookie)
* timestamp: the timestamp of the action

Other:
* Net Promoter Score (NPS) survey submissions (see the survey here)
* Chrome Uninstall Feedback

###Marketing

We track our lead activity using HubSpot, which we use for marketing automation and our CRM.

Main HubSpot events:
* Request to demo form
* Contact form
* In-product trial request: A user installs an instance on a local machine and checks the box that requests an enterprise trial
* Create a Sourcegraph.com account
* Chrome uninstall feedback: feedback is provided when the chrome extension is uninstalled
* Private instance: A private instance was created on a user’s local machine AND they created an account to use the product (note: users can’t use Sourcegraph on a local machine until they create an account so installs =! accounts)
* Sourcegraph.com Account: A user creates an account while using the Sourcegraph-hosted instance
* NPS submission: User submits an [NPS survey](https://sourcegraph.com/survey) and includes their email
* Live conferences and events: Event leads and live-blog subscribers=

## Using Looker

Coming soon
