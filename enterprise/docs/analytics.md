# Event logging how-to

## Why track frontend events?

User data give us the best way to track business progress, measure success of our individual projects, and improve our users' experiences.

If you ever have questions about why or how to track events, ask [Dan](mailto:dan@sourcegraph.com) to grab a coffee!

## Event logging 101

The goal of the `eventLogger` API is to be as simple as possible. If you ever wonder if it's the right call to log something — just do it!

To log an event on a "Find References" action (e.g. a button being clicked):

1.  Add the following import to your file (with the appropriate path):

```ts
import { eventLogger } from '../tracking/eventLogger'
```

2.  Call:

```ts
eventLogger.log('FindReferencesButtonClicked')
```

3.  Turn on event debug mode in your browser to confirm it's working as expected. In your browser's console, type: `localStorage.eventLogDebug="true"`. After this, make sure your console is displaying `All levels` of output, and not filtering debug messages. If you see the text: "X items hidden by filters," you likely need to select "Verbose" output using the dropdown menu next to it.

You should begin to see events stream into your console as you do actions.
<BR><BR>
When you click your new "Find References" button, you should see the following appear in gray in the console:

```
EVENT FindReferencesButtonClicked
```

4.  Test usage, and confirm that the event only occurs when you want it to. React's lifecycle can make this tricky — for example, never run `eventLogger.log(...)` directly inside of a `render` method!

That's it! Once your commit is deployed, events will begin being logged to our production BigQuery data store.

## Logging custom event data

If you want to log custom data with an event, `eventLogger.log` accepts an optional second parameter: `eventProps`. This parameter should be an object with named properties. For a real-world example, when a user clicks a button to install our Chrome extension (versus our Firefox extension), the event call looks like:

```ts
eventLogger.log('BrowserExtInstallClicked', {
  marketing: {
    browser: 'Chrome',
  },
})
```

Note that our event logger automatically adds contextual and persistent data to every event — e.g. an identifier for the current user who did the action, the current URL, the current timestamp, etc. Most events can be safely logged without any custom data.

### Namespacing

You may have noticed in the example above that the `browser` property falls under a `marketing` namespace. This helps prevent identically-named properties used in different contexts from conflicting.

In general, try to re-use existing namespaces where relevant. If you're developing a new product feature/UX, add a new _descriptive_ namespace for it. Ask Dan for help with namespacing if you plan to log custom properties!

## Pageview events

You may have seen another category of events in the Sourcegraph codebase — pageviews. Unlike most events, which occur on specific user actions, pageview events should be passively logged every time a user views a page.

They are executed in a similar way, but using the `logViewEvent` method, e.g.:

```ts
eventLogger.logViewEvent('SearchResults')
```

You must take care to only log these on the initial page load. This can be tricky in React — these should NEVER appear in `render` methods. Generally they will only be logged in `ComponentDidMount` (and occasionally in `ComponentWillReceiveProps`).

## Seeing results

There are two ways to view event data. If you prefer to run raw SQL, you can get access to our [BigQuery instance](https://bigquery.cloud.google.com/dataset/telligentsourcegraph:telligent). If you prefer to use a visualization/pivot-table tool, use [Sourcegraph's Looker instance](https://sourcegraph.looker.com).

More to come on how to use these tools. For now, ask [Dan](mailto:dan@sourcegraph.com) for more information!

## The messy details: the data pipeline overview

> Rather than waiting for events to show up in BigQuery or Looker to validate that your events are being logged properly, use the `localStorage.eventLogDebug` flag. The ETL jobs that do data cleaning and stitching on the backend only run every 3 hours, so there will be a delay.

Sourcegraph uses the [Telligent](https://github.com/telligent-data/telligent-javascript-tracker) tracker library (based on the popular [Snowplow](https://github.com/snowplow/snowplow) tracker) to manage logistics for tracking, such as user sessions and user cookies, data enrichment, and network requests.

Events are sent in real-time, without batching, and make their way through a de-duplication and enrichment pipeline. They end up batched in json files in a Google Cloud Storage bucket, from where they get loaded into BigQuery in real-time.

These real-time events get loaded into the `telligent.events` table in BigQuery, which contains raw event data.

Every 3 hours, a series of [Sourcegraph-managed ETL jobs](https://github.com/KattMingMing/SGMetricsPipeline) run, enhancing the raw data. These jobs:

- Filter out junk data (from bots, frontend bugs, and more).
- Analyze all events to stitch together a consistent concept of a user, across sessions, products (e.g. web vs. browser extensions), and devices. This is based on Sourcegraph.com user profiles, where possible.
- Perform higher-level analyses through data aggregation.

These jobs generate a series of tables that can be accessed and analyzed, including:

- `events_users`: a clone of the raw events table, but with junk data filtered, and user-level properties (i.e., in addition to the event-level properties) appended to each row
- `users`: a smaller table with a row for each uniquely identified user
- `sessions`: a table with a row for each user session
- `daily_engagement`: a table with a row for each day-user pair, used for x/7 DWAU calculations
- `orgs` and `memberships`: tables for linking users to Sourcegraph organizations
- and more...
