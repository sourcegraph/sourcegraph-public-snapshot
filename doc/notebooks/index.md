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

# Notebooks
Notebooks enable powerful live–and persistent–documentation, shareable with your organization or the world. As of release 3.39, Notebooks are Generally Available (GA) on Sourcegraph Cloud and in Sourcegraph enterprise installs. You can explore all the public notebooks on Sourcegraph.com without an account, or create a [Sourcegraph Cloud](https://about.sourcegraph.com/get-started/cloud) account to start creating your own Notebooks. If you're a Sourcegraph enterprise user and are on version 3.39 or later, you can find a Notebooks item in the header navigation, or by going to https://<your-internal-sourcegraph-url>/notebooks.

Inspired by Jupyter Notebooks and powered by Markdown and Sourcegraph's code search, Notebooks let you and your team create living documentation that interacts directly with your code. You can leverage Notebooks to onboard a new teammate, document [known vulnerabilities](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MQ==), a [common pattern](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6OTI=) in your codebase, or [useful Sourcegraph queries](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MTU=).

![](https://storage.googleapis.com/sourcegraph-assets/docs/images/notebooks/notebooks_home.gif)

Notebooks have powerful content creation features, like a [notepad](../notebooks/notepad.md) and multiple block types, each with their own unique capabilities. If you're familiar with Jupyter Notebooks, then you already understand the blocks concept. You can add as many of each block as you want to Sourcegraph's notebooks.

# Notebook blocks
The currently supported block types are:
- Query
- File
- Symbol
- Markdown

[Read more about block types](../notebooks/blocks.md).

# Searching
Notebooks created through the web interface are full text searchable from the /notebooks page. Each tab has its own search box and each search box is scoped to that tab. The exception is the Explore tab, which searches all notebooks you have access to.

Searches will match on notebook titles and any text in Markdown blocks. Searching through other block types is not supported because they are dynamic in nature.


## Accessing Notebooks in older versions of Sourcegraph
In versions older than 3.39 (beginning in 3.36) Notebooks are behind an experimental feature flag. If you're running versions 3.36-3.38 and want to try out Notebooks, enable them in global settings:

```
"experimentalFeatures": {
    "showSearchNotebook": true
}
```

<div class="cta-group">
  <a class="btn btn-primary" href="quickstart">★ Quickstart</a>
</div>

## Explanations
- [Sharing notebooks](../notebooks/notebook-sharing.md)
- [Embedding notebooks](../notebooks/notebook-embedding.md)
- [The notepad](../notebooks/notepad.md)
- [Block types](../notebooks/blocks.md)
