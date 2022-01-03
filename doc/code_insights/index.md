# Code Insights <a href="../admin/beta_and_experimental_features"><sup><span class="badge badge-beta">Beta</span></sup></a>

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
  background-image: url(code_monitoring/file-icon.svg);
  margin-right: 0.5em;
  margin-bottom: -0.29em;
}

body.theme-dark .markdown-body ul li:before {
  filter: invert(50%);
}

</style>

<p class="subtitle">Anything you can search, you can track and analyze</p>

<img src="https://sourcegraphstatic.com/docs/images/code_insights/insights_index_light.png" class="screenshot theme-light-only" />
<img src="https://sourcegraphstatic.com/docs/images/code_insights/insights_index_dark.png" class="screenshot theme-dark-only" />

<p class="lead">Code Insights reveals high-level information about your codebase, based on both how your code changes over time and its current state.</p>

Code Insights is based on our universal code search, making it precise and configurable. Track anything that can be expressed with a Sourcegraph search query: migrations, package use, version adoption, code smells, codebase size and much more, across 1,000s of repositories.

### Code Insights is in Beta

Code Insights are currently in [Beta](../admin/beta_and_experimental_features.md). If you want early access or support, please [send us an email](mailto:feedback@sourcegraph.com).

> NOTE: While in Beta, Code Insights is free for enterprise customers. Once Code Insights is generally available and no sooner than January 1, 2022, to continue using Code Insights may require a separate paid plan.

We're still polishing Code Insights and you might find bugs while we’re in beta. Please [share any bugs 🐛 or feedback](mailto:feedback@sourcegraph.com) you have to help us make Code Insights better.

Code insights docs are actively in progress during the beta.

<div class="cta-group">
<a class="btn btn-primary" href="quickstart">★ Quickstart</a>
<a class="btn" href="language_insight_quickstart">Language Insight Quickstart</a>
</div>

## [Explanations](explanations/index.md)

- [Administration and security of Code Insights](explanations/administration_and_security_of_code_insights.md)
- [Automatically generated data series for version or pattern tracking](explanations/automatically_generated_data_series.md)
- [Code Insights filters](explanations/code_insights_filters.md)
- [Current limitations of Code Insights (Beta limitations)](explanations/current_limitations_of_code_insights.md)
- [Viewing code insights](explanations/viewing_code_insights.md)

## [How-tos](how-tos/index.md)

- [Creating a dashboard of code insights](how-tos/creating_a_custom_dashboard_of_code_insights.md)
- [Filtering an insight](how-tos/filtering_an_insight.md)
- [Troubleshooting](how-tos/Troubleshooting.md)

## [References](references/index.md)

- [Common use cases and recipes](references/common_use_cases.md)
- [Common reasons code insights may not match search results](references/common_reasons_code_insights_may_not_match_search_results.md)
