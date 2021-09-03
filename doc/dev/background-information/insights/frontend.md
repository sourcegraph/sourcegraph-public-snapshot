# Code Insights FE architecture overview

## Table of contents

- [Insights directory structure](#insights_directory_structure)
- [Insight types](#insight_types)
- [Insight configuration storing](#insight-configuration-storing)
- [Quick intro to the setting cascade](#quick-intro-to-the-setting-cascade)
- [Insight consumers](#insight-consumers)
  - [The dashboard page](#the-dashboard-page) 
  - [The directory and homepage](#the-directory-page) 
- [Code Insights loading logic in details](#code-insights-loading-logic-in-details)


## Insights directory structure

We store all insights related parts of components and logic in the `insights` directory.
The full path to this folder is `./client/web/src/insights`. There you can find all components
and code insights shared logic.

This directory has the following structure 

- `/components` - all shared and reusable components for code insights pages and others components.
- `/core` - backend-related logic such as `InsightsApiContext` and data fetchers. Also, analytics and core interfaces for code insights entities.
- `/hooks` - all shared across insights components hook-based logic.
- `/mocks` - mock collections for unit tests and storybook stories.
- `/pages` - all pages-like code insights components (such as `DashboardPage` or `InsightCreationPage`)
- `/utils` - common helpers 
- `insight-global-styles.css` All code insight related non-css-module styles that must be included into the main css bundle
  via scss `@import` in `SourcegraphWebApp.scss` file.
- `InsightsRouter.tsx` - The main entry point for all code-insights pages.


## Insight types 

At the moment, we have two different types of insights.

1. **Extension based insights.** <br/>
These types of insights fetch data via Extension API. At the moment, we have at least two 
insight extensions that work this way.
<br /> &nbsp;
   - [Search based insight (line chart)](https://github.com/sourcegraph/sourcegraph-search-insights)
   - [Code stats insight (pie chart)](https://github.com/sourcegraph/sourcegraph-code-stats-insights). <br />
   
> These extensions are running on the frontend and making multiple network requests to prepare and process
insight data and then pass this data to the React page component to render charts.
> You can find extension documentation here [Sourcegraph extensions](../sourcegraph_extensions.md)

2. **Backend based insights.** <br/>
These insights are working via our graphql API only. At the moment, only search based insights (line chart)
can be backend-based. Code stats insights (pie chart) only work via extension API.

You can find typescript types that describe these insight entities
in [/core/types/insights/index.ts](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/core/types/insight/index.ts)

## Insight configuration storing

To be able to fetch data for an insight, we need to store some insight configuration (input data), such as
data series config - line color, search query string, the title of data series.

We use the setting cascade to store these insight configurations. For example, Search based insight configuration looks 
something like this 

```json
{
  // other setting cascade subject properties
  // ...
  // ...

  "searchInsight.insight.someInsight": {
    "title": "Some insight",
    "repositories": ["github.com/test/test", "github.com/sourcegraph/sourcegraph"],
    "series": [
      { "title": "#1 data series", "query": "test", "stroke": "#000" },
      { "title": "#2 data series", "query": "test2", "stroke": "red" }
    ]
  } 
}
```

Search based insight extension [takes this configuration](https://github.com/sourcegraph/sourcegraph-search-insights/blob/master/src/search-insights.ts#L47-L56)
and run network requests according to this information.

_Example from the search based insight codebase._

```ts
const settings = from(sourcegraph.configuration).pipe(
        startWith(null),
        map(() => sourcegraph.configuration.get().value)
    )

const insightConfigs = settings.pipe(
    map(
        settings =>
            Object.entries(settings)
              .filter(([key]) => key.startsWith('searchInsights.insight.')) as [
                string,
                Insight | null | false
            ][]
    )
)
```

> Code stats extension insight does exactly this thing to get insight configurations and get insights data.

Backend based insights also have their insight configurations, and they are also stored
in the same settings cascade but by special property key `insights.allrepos`

```json

{
  // other setting cascade subject properties
  // ...
  // ...
  
  "insights.allrepos": {

    "searchInsight.insight.someInsight": {
      "title": "Some insight",
      "series": [
        { 
          "title": "#1 data series",
          "query": "test", 
          "stroke": "#000"
        }
      ]
    }
  }
}
```

You can find typescript types that describe these insight entities 
in [/core/types/insights/index.ts](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/core/types/insight/index.ts)

> This way to store insights isn't the best, but this is the easiest way to get insight configs from extensions.
> Eventually, we want to migrate all insights to our BE and store them in real DB.

We also can write to these `jsonc` files and therefore create or delete some insights. This is actually how it works
on the creation UI. After submitting, we just produce a valid insight configuration, write this config to the setting subject `jsonc` file,
and upload this new config back via GQL API.


**Important note:** Each insight (search backend or extension based or code stats insight) has the 
visibility setting.

![insight-visibility](https://storage.googleapis.com/sourcegraph-assets/code_insights/insight-visibility.png)

This setting is responsible for storing the insight in some particular setting subject file (personal, org level, or global jsonc file)
For example, if I created insight with some particular organization, the logic behind the creation page will load
`jsonc` file of organization subject then add newly created insight (its configuration) to this `jsonc` file
and then saves it via our gql API and trigger re-hydration for local settings, so you don't need to reload the page
to see the last updated settings cascade on the page.

Also, this setting affects what dashboard will be used to show this insight.

It is worth mentioning that we use setting cascade subjects not only for storing insight configurations but also dashboard configurations.
We will cover this in other sections further.


## Quick intro to the setting cascade

As we mentioned before all insights (their configurations) are stored in the setting cascade. But what is the setting cascade?

In a nutshell, this is just a system around a couple of configuration files (called subjects). These files
are just `jsonc` files. But each subject (`jsonc`) file has its cascade level (means that setting cascade has some file hierarchy)

![settings-cascade-levels.svg](assets/settings-cascade-levels.svg)

So eventually, the FE merges all these files in one big `jsonc` object and deserializes this object to a common js object.
You can find this merge logic here [/client/shared/src/settings/settings.ts](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/shared/src/settings/settings.ts) 
`mergeSettings` function.

We use settings cascade a lot in different places, In fact, our dashboard system and insight visibility were built on top
of settings cascade levels.  


## Insight consumers 

The first thing we should know before diving deep into the code insights codebase are the places
where we're rendering/showing code insights component.

At the moment, we have at least three different places where you can find
code insights components/logic

1. The dashboard page
   - The dashboard page itself 
   - Creation UI pages (search and code stats insights)
3. The Home (search) page
4. The directory page

Further, we will cover all three places where we use something from the insights code-base.


### The dashboard page

![Dashboard select page](https://storage.googleapis.com/sourcegraph-assets/code_insights/dashboard-select-page.png)

This page is kind of the main source of insights in the app at the moment. By the dashboard page, you can find any
accessible to you insights by going through different dashboards.

As you can see on the screenshot above, the page has the select component to pick the right dashboard.

All dashboards have the following hierarchy

1. **All insights dashboard**
<br/> This dashboard contains all available user insights from all places.
All users have this dashboard by default, users can't delete this dashboard.
2. **< user name >'s insights** 
<br/> This is something called a **built-in dashboard**. This dashboard represents
your personal level of settings, which means that this dashboard contains all insights from your
personal setting cascade subject `jsonc` file and only these insights. This dashboard also can't be deleted.
3. **< organization name >'s insights**
<br /> This is also a built-in dashboard. It has the same functionality as a personal dashboard but contains
insights only from this org level jsonc setting subject file.
4. **Global insights**
<br/> Also, built-in dashboard. This dashboard contains insights from the subject of the global setting,
which is shared across all users within single sourcegraph instance. You should be an admin to be able to
write and update this setting subject.

All three types of dashboards (user, organizations, and global) can have their **custom dashboards** via dashboard creation UI.

So this fact leads us that we need to store dashboard configuration somewhere. To be able to filter some insight and leave only
those insights which were added/included to some particular dashboard.

We use setting cascade subject to store dashboard configurations. 

**Example of this dashboard config** 

```json
{
  
  "insights.dashboards": {
    "myPersonalDashboard": {
      "id": "2e20a79f-d32d-4433-b367-c0874d391e78",
      "title": "My personal dashboard",
      "insightIds": [
        "searchInsights.insight.allReposInsightTesting",
        "searchInsights.insight.stringTestForNamesFelix"
      ]
    },
    "myPersonalDashboard2": {
      "id": "4d609411-1586-48e9-8ed4-ecac80aff11f",
      "title": "My personal dashboard 2",
      "insightIds": [
        "searchInsights.insight.test13QueryJesteRepoExpectedPresentDay119KResuls"
      ]
    }
    // Other dashboard configs
  }
}
```

You can find dashboard typescript types for these dashboard properties in [core/types/dashboard/index.ts](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/core/types/dashboard/index.ts)

Let's take a look at the dashboard system in action. For example, let's describe what will happen when we go to the `/insights/dashboard/<personal subject id>`

1. We extract the dashboard id from the URL in the `DashboardPage` component via react-router URL options.
2. With `useDashboard` hook ([source link](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/hooks/use-dashboards/use-dashboards.ts)) we select/extract all 
reachable dashboards from all settings cascade levels.
3. Then we map the dashboard id from the URL and all dashboard configs, extract information about dashboard 
like insights ids (`insightIds` property)
4. Pass `insightsId` information to component for rendering insights (in case of the dashboard page this component 
is `SmartInsightsViewGrid.tsx` [source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/components/insights-view-grid/SmartInsightsViewGrid.tsx)) 
5. `SmartInsightsViewGrid.tsx` component will iterate over all `insightIds` get insight configuration from setting
cascade by `useInsight()` [source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/hooks/use-insight/use-insight.ts) hook and pick the right component 
(either `BackendInsight.tsx` [source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/components/insights-view-grid/components/backend-insight/BackendInsight.tsx) 
or `ExtensionInsight.tsx` [source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/components/insights-view-grid/components/extension-insight/ExtensionInsight.tsx))
6. Then this backend or extension insight component will load insight data either by GQL API in the case of Backend Insight or 
by Extension API in case of extension insight. 

![insight-dashboard-loading.svg](./assets/insight-dashboard-loading.svg)

> **Note** We load our insights one by one with a maximum of two insight data requests in parallel to avoid HTTP request bombarding and HTTP 1 limit 
> with only six requests in parallel. To do that, we use `useParallelRequests` react hook [source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/hooks/use-parallel-requests/use-parallel-request.ts)


### The directory page

![directory-page.png](https://storage.googleapis.com/sourcegraph-assets/code_insights/directory-page.png)

The directory page is another place where you can find insights (and other extension-based things).
This page renders an insight grid component with all insights that you have in your subject settings so that we could say that
this is kind of analog of the All insights dashboard on the dashboard page.

But this page uses a slightly different approach how to load insights data. If on the dashboard page
insights card components were responsible for fetching logic, then on the directory page, the component that renders
the insight grid component (in this case, directory pages renders the `StaticInsightsViewGrid` [source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/components/insights-view-grid/StaticInsightsViewGrid.tsx))
this component has to load insight data on its own. In fact, the **home (search) pages use exactly the same approach** to fetch and display insight data.


## Code Insights loading logic in details

All async operation which is related to fetching data or calling something from the extension API is produced and provided via
React Context system. Code Insights API is accessible via `InsightAPIContext` [source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/core/backend/insights-api.ts)

This was done in this way to mock and change the implementation of async (backend API or extension API) calls in unit tests. 

Let's take a look on simple version of `ExtensionInsight` component


```tsx
  function ExtensionInsight(props) {
  const { viewId } = props
  const { getExtensionViewById } = useContext(InsightsApiContext)
  
  const { data, loading } = useParallelRequests(
    useMemo(() => () => getExtensionViewById(viewId), [viewId])
  )
  
  return (/* Render insight chart */)
}
```

So in this component we use `getExtensionViewById` function from our `InsightsApiContext` context. If we go to `InsightsApiContext`
[source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/core/backend/insights-api.ts) definition we will see that this is just an object with some async function collection. 
All these functions and their interfaces are described in this one interface `ApiService` [source](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/insights/core/backend/types.ts)

Then if we want to write some unit test for the `ExtensionInsight` component we will write something like this

```tsx
const mockAPI = createMockInsightAPI({
  getExtensionViewById: () => ({ /* Some mock insight data */})
})

it('ExtensionInsight should render', () => {
  
  const { getByRole } = render(
    <InsightsApiContext.Provider value={mockAPI}>
      <ExtensionInsight viewId="TEST_VIEW_ID" />
    </InsightsApiContext.Provider>
  )
  
  /* Further test logic here */
})
```


