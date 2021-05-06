# Code Insights

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
  background-image: url(batch_changes/file-icon.svg);
  margin-right: 0.5em;
  margin-bottom: -0.29em;
}

body.theme-dark .markdown-body ul li:before {
  filter: invert(50%);
}

</style>

<p class="subtitle">Track migrations, analyze trends, and understand where code gets used within your codebase</p>

<p class="lead">
Code Insights turn your codebase into a database: you can track over time the use of specific libraries or functions, the removal of vulnerable or deprecated code snippets, or a migration to a new service. You can also explore metadata like language usage or test coverage across groups of repositories. Then, control where insights display and whether or not others in your organization can also view them. 
</p>

<p>[TODO ADD CODE INSIGHTS IMAGE]</p>

<div class="cta-group">
<a class="btn btn-primary" href="quickstart">★ Quickstart</a>
<a class="btn" href="explanations/introduction_to_code_insights">Introduction to Code Insights</a>
<a class="btn" href="references/requirements">TODO</a>
</div>

## Getting started

<div class="getting-started">
  <a href="TODO" class="btn" alt="Run through the Quickstart guide">
   <span>New to Code Insights?</span>
   </br>
   Run through the <b>quickstart guide</b> and create a code insight in less than 2 minutes.
  </a>

  <a href="TODO" class="btn" alt="TODO">
   <span>Demo video</span>
   </br>
   TODO, doesn't have to be demo.
  </a>

  <a href="TODO" class="btn" alt="TODO">
   <span>TODO OVERVIEw</span>
   </br>
   Find out what Code Insights is, learn key concepts and see what others use them for.
  </a>
</div>

## Explanations

- [Introduction to Code Insights]
- [Types of Code Insights]
- [User viewing permissions of Code Insights]
- [Administration and Security of Code Insights]
- [How Code Insights work]

## How-tos

- Create or remove a code insight
    - [Create a search-based code insight]
    - [Create a language usage code insight]
    - [Create an extension-provided code insight]
    - [Create your own extension that adds a code insight]
    - [Delete a code insight]
- Control how code insights display
    - [Control where code insights appear]
    - [Control who can see your code insights]

## Tutorials

- [Tracking migration from React Functional to Class Components with a search-based insight]
- [Tracking database usage using a search-based insight]
- [Tracking code removals with a search-based insight]
- [Viewing languages on the core Sourcegraph repository with a language insight]
- [Viewing code coverage insights using the Sourcegraph CodeCov extension]
- [Full set of common examples] 

## References

- [Code insights source code]
- [Troubleshooting]
- [Code Insights public roadmap]
