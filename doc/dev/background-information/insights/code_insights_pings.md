# Code Insights Pings

Code Insights pings allow us to quantitatively measure the usage and success of Code Insights. This page is a source of truth for detailed explanations, statuses, and implementations of our pings. 

We keep this docs page up to date because pings are a vital component of our product knowledge and prioritization process, and a broken or incorrect ping impacts 3-5 months of data (because that's how long a fix takes to propagate). 

## Metrics

### Additions count, edits count, and removals count 

**Intended purpose:** To track how many times customers have created, edited, and removed insights, by week. 

**Functional implementation:** The current implementation of Code Insights stores insight configurations in settings files. This ping works by diffing settings files if users edit their settings in the setting editor at the setting edit page.
Also, we track insight creating/editing/deleting events in the creation UI form and insight context menu component with standard telemetry service calls.

**Other considerations:** This is an "imperfect" ping because not all additions + removals directly translate to a new insight or a deleted insight, due to the complications with using settings files as a source of truth. We'll be fixing this when we migrate to a backend database. Note also we're using this as a "total insights" metric for the same imperfect reason (additions - removals = total created) and when we migrate to the backend database we should build an additional separate ping that is just "total insights existing on the instance" per week. 

- Aggregation: By week 
- Event Code: [InsightAddition](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%27InsightAddition%27&patternType=literal), [InsightEdit](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%27InsightEdit%27&patternType=literal), [InsightRemoval](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%27InsightRemoval%27&patternType=literal)
- PRs: [#17805](https://github.com/sourcegraph/sourcegraph/pull/17805/files)
- **Version Added:** 3.25
- **Version(s) broken:**  3.31-3.35.0 (does not count backend insights) ([fix PR](https://github.com/sourcegraph/sourcegraph/pull/25317))


### Hovers count

**Intended purpose:** To track how many times users hover over a datapoint to see the tooltip on the graph, or "dig in" to the information. 

**Functional implementation:** This ping works by firing an event on the client when a user hovers over a datapoint on a code insight. 

**Other considerations:** N/A

- Aggregation: By week 
- Event Code: [InsightHover](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%27InsightHover%27&patternType=literal) 
- PRs: [#17805](https://github.com/sourcegraph/sourcegraph/pull/17805/files)
- **Version Added:** 3.25
<!-- - **Known versions broken:** N/A -->

### UI customizations count

**Intended purpose:** To track how many times users resize the insight graphs. 

**Functional implementation:** This ping works by firing an event on the client when a user resizes a Code Insights graph on the page. 

**Other considerations:** N/A

- Aggregation: By week 
- Event Code: [InsightUICustomization](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%27InsightUICustomization%27&patternType=literal) 
- PRs: [#17805](https://github.com/sourcegraph/sourcegraph/pull/17805/files)
- **Version Added:** 3.25
<!-- - **Known versions broken:** N/A -->

### Data point clicks count

**Intended purpose:** To track how many times users click a datapoint to get to a diff search. 

**Functional implementation:** This ping works by firing an event on the client when a user clicks an individual data point of an insight graph, which takes them to a diff search. 

**Other considerations:** Because this functionality does not yet exist for backend insights, it only tracks clicks on frontend insights. 

- Aggregation: By week 
- Event Code: [InsightDataPointClick](
https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%27InsightDataPointClick%27&patternType=literal) 
- PRs: [#17805](https://github.com/sourcegraph/sourcegraph/pull/17805/files)
- **Version Added:** 3.25
<!-- - **Known versions broken:** N/A -->

### Page views count

**Intended purpose:** To track how many times users view insights pages. 

**Functional implementation:** This ping works by firing an event on the client when a user views _any_ /insights page, whether it's creating or viewing insights.  

**Other considerations:** As we add new insights pages it's important to make sure we're adding pages to this counter. 

- Aggregation: By week 
- Event Code: [InsightsPageView](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+InsightsPageView&patternType=regexp) 
- PRs: [#17805](https://github.com/sourcegraph/sourcegraph/pull/17805/files)
- **Version Added:** 3.25
- **Version(s) broken:** 3.25-3.26 (not weekly)([fix PR](https://github.com/sourcegraph/sourcegraph/pull/20070/files)), 3.30 (broken when switching to dashboard pages, didn't track dashboard views)([fix PR](https://github.com/sourcegraph/sourcegraph/pull/24129/files))

### Unique page views count

**Intended purpose:** To track how many unique users are viewing insights pages each week. 

**Functional implementation:** This ping works by firing an event on the client when a unique user views _any_ /insights page for the first time that week, whether it's creating or viewing insights.  

**Other considerations:** As we add new insights pages it's important to make sure we're adding pages to this counter. 

- Aggregation: By week 
- Event Code: [InsightsUniquePageView](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+InsightsUniquePageView&patternType=regexp) 
- PRs: [#17805](https://github.com/sourcegraph/sourcegraph/pull/17805/files)
- **Version Added:** 3.25
- **Version(s) broken:** 3.25-3.26 (not weekly)([fix PR](https://github.com/sourcegraph/sourcegraph/pull/20070/files)), 3.30 (broken when switching to dashboard pages, didn't track dashboard views)([fix PR](https://github.com/sourcegraph/sourcegraph/pull/24129/files))

### Org-visible insights count (Total) 

**Intended purpose:** To track how many insights are visible by more than just the creator of the insight. 

**Functional implementation:** we gather this on the backend from TODO what table maybe?

**Other considerations:** N/A

- Aggregation: total time, by insight type
- Event Code: [InsightOrgVisible](https://sourcegraph.com/search?q=context:global+insightorgvisible+r:sourcegraph/sourcegraph%24&patternType=literal)
- PRs: [#21671](https://github.com/sourcegraph/sourcegraph/pull/21671/files)
- **Version Added:** 3.29
- **Version(s) broken:** 3.31-3.35.0 (doesn't handle backend insights) [fix PR](https://github.com/sourcegraph/sourcegraph/pull/28425)

### First time insight creators count

**Intended purpose:** To track the week and count of the first time a user(s) creates a code insight, of any type, on an instance. The sum of first time insight creators count over all time is equal to the total number of unique creators who have made an insight.

**Functional implementation:** This metric queries the insight table for new addition events, then filters by unique IDs that appeared for the first time that week. 

**Other considerations:** TODO does this ping include creators who create via the API? 

- Aggregation: By week
- Event Code: [WeeklyFirstTimeInsightCreators](https://sourcegraph.com/search?q=context:global+WeeklyFirstTimeInsightCreators+r:sourcegraph/sourcegraph%24&patternType=regexp)
- **Version Added:** 3.25
- **Version(s) broken:** 3.31-3.35.0 (doesn't handle backend insights, other bugs)

### Total count of insights grouped by step size (days)

**Intended purpose:** To track the x-axis (time window) set by users on frontend insights, to help prioritize features related to setting time windows. 

**Functional implementation:** this metric runs on the backend over all the insights. 

**Other considerations:** N/A

- Aggregation: total 
- Event Code: [GetTimeStepCount](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+GetTimeStepCounts&patternType=literal)
- **Version added:** 3.29
- **Version(s) broken:** 3.31-3.35.0 [fix PR](https://github.com/sourcegraph/sourcegraph/pull/28425)

### Code Insights View/Click Creation Funnels

**Intended purpose:** These pings allow us to both understand how the view/click/view/click conversion funnel works for the creation flows of all existing types of insights, as well as smell-check other pings. The reason we use both "view" and "button clicks" in this funnel is that it's possible to view a page without "funneling through" via the prior page's CTA (for example: you can reach the creation/edit screen by the "edit" button, which does not involve logging a click on the "create search insight" button).

**Functional implementation:** These events fire on the frontend when the user takes the below actions. 

**Other considerations:** 

- Aggregation: By week
- Event Code:
   - For the "search insight" funnel: (1) [ViewCodeInsightsCreationPage](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+ViewCodeInsightsCreationPage&patternType=regexp), (2) [CodeInsightCreateSearchBasedInsightClick](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+CodeInsightsCreateSearchBasedInsightClick&patternType=regexp), (3) [ViewSearchBasedCreationPage](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+SearchBasedCreationPage&patternType=regexp), (4.1) [SearchBasedCreationPageSubmitClick](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+SearchBasedCreationPageSubmit&patternType=regexp) OR (4.2) [SearchBasedCreationPageCancelClick](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+SearchBasedCreationPageCancelClick&patternType=regexp)
   - For the "language stats insight" funnel: (1) [ViewCodeInsightsCreationPage](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+ViewCodeInsightsCreationPage&patternType=regexp), (2) [CodeInsightsCreateCodeStatsInsightClick](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+CodeInsightsCreateCodeStatsInsightClick&patternType=regexp), (3) [ViewCodeInsightsCodeStatsCreationPage](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+CodeInsightsCodeStatsCreationPage&patternType=regexp), (4.1) [CodeInsightsCodeStatsCreationPageSubmitClick](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+CodeInsightsCodeStatsCreationPageSubmitClick&patternType=regexp) OR (4.2) [CodeInsightsCodeStatsCreationPageCancelClick](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+CodeInsightsCodeStatsCreationPageCancelClick&patternType=regexp)
   - For the "extensions insight" funnel: (1) [ViewCodeInsightsCreationPage](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+ViewCodeInsightsCreationPage&patternType=regexp), (2) [CodeInsightsExploreInsightExtensionsClick](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+CodeInsightsExploreInsightExtensionsClick&patternType=regexp). 
- **Version added:** 3.29
<!-- - **Version(s) broken:**  -->
