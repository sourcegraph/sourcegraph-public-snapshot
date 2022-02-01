# Viewing Code Insights

Code Insights can display in 3 areas of Sourcegraph:

- On the insight dashboard pages, which begin with `/insights/dashboards`; enabled by default
- On repository and directly pages for insights that run over that repository; enabled by default
- On the search home page, at `/search`; disabled by default

## Insights dashboards

The main way to view code insights is on a dashboard page. Dashboards have a unique name and visibility level.

There are three possible visibility levels:

- Private: visible only to you
- Shared with [an organization](../../../admin/organizations.md): visible to everyone in the organization
- Global: visible to everyone on the Sourcegraph instance

### Built-in dashboards

A default "dashboard" of all insights visible to a user appears as the homepage for the Insights navigation bar item. 

To add insights to your own custom dashboard, see [Creating a custom dashboard of code insights](../how-tos/creating_a_custom_dashboard_of_code_insights.md).

### Dashboard visibility

A Dashboard's visibility level owns the visibility of code insights on the dashboard. If you add an Insight that was on a private dashboard to an organization or global dashboard, now people with access to that dashboard can view the insight. When you first create an insight, it defaults to appearing on the dashboard from which you clicked the "create" button. If you create an insight directly on the create page or from the default dashboard, it will appear on the default dashboard. 

### Insights still enforce individual permissions regardless of dashboard visibility

The dashboard visibility levels have no impact on the data in the insight itself that is displayed to an individual user: insights [enforce the viewer's permissions](administration_and_security_of_code_insights.md#code-insights-enforce-user-permissions).

This means that two organization users with different repo read permission sets might see different values for the same insights on the same dashboards, if it contains results from a repository only one user can view.

(This also means that if you change a private-visible dashboard with both organization-visible and private-visible insights so that the dashboard is now visible to an organization, then the organization can only see the organization-visible insights on that dashboard. This is non-optimal and a bit awkward to convey, and we are actively improving this UX. If you have strong thoughts, please do [leave them on the issue](https://github.com/sourcegraph/sourcegraph/issues/23003).)

### Insights can be on multiple dashboards

You can attach insights to multiple dashboards.

## Repository and directory pages

On repository pages, any code insight that runs over that repository can display. On directory pages, code insights defined for that repository can display **and** run only over files that are children of the directory.

This is disabled by default as it is the not the primary way to view insights and there are (yet) no ways to select only showing a subset of possible insights, but it can be enabled with a flag in your global, organization, or user settings:

```json
"insights.displayLocation.directory": true
```

## Search home page

Code insights can display below the search bar on the search home page. This is disabled by default because we have not yet built the capability to select which insights display on the search home page (so they all do). This can be enabled with a flag in your global, organization, or user settings:

```json
"insights.displayLocation.directory": true
```

