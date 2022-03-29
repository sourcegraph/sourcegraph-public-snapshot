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
Notebooks enable powerful live–and persistent–documentation, shareable with your organization or the world. Notebooks are currently in public [Beta](https://sourcegraph.com/notebooks?order=stars-desc) on Sourcegraph Cloud and in Sourcegraph enterprise installs at version 3.36 or later. You can explore all the public notebooks on Sourcegraph.com without an account, or create a [Sourcegraph Cloud](https://about.sourcegraph.com/get-started/cloud) account to start creating your own Notebooks.

Inspired by Jupyter Notebooks and powered by Markdown and Sourcegraph's code search, Notebooks let you and your team create living documentation that interacts directly with your code. You can leverage Notebooks to onboard a new teammate, document [known vulnerabilities](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MQ==), a [common pattern](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6OTI=) in your codebase, or [useful Sourcegraph queries](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MTU=).

<!-- Notebooks image TODO: get uploaded to GCP -->
![](https://storage.googleapis.com/sourcegraph-assets/docs/images/notebooks/notebooks_home.gif)

## Notebooks are in Beta
Notebooks are currently in public [Beta](https://sourcegraph.com/notebooks?order=stars-desc) on Sourcegraph Cloud and in Sourcegraph enterprise installs at version 3.36 or later. You can explore all the public notebooks on Sourcegraph.com without an account, or create a [Sourcegraph Cloud](https://about.sourcegraph.com/get-started/cloud) account to start creating your own Notebooks.

To try out notebooks on your enterprise install, enable them in global settings:

```
"experimentalFeatures": {
    "showSearchNotebook": true
}
```

We're still actively developing Notebooks while in Beta and we'd love your [feedback and bugs](mailto:feedback@sourcegraph.com) so we can make them better.

<div class="cta-group">
  <a class="btn btn-primary" href="quickstart">★ Quickstart</a>
</div>

## Explanations
- [Sharing Notebooks](../notebooks/notebook-sharing.md)
- [Embedding Notebooks](../notebooks/notebook-embedding.md)
