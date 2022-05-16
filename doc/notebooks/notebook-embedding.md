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
# Embedding Notebooks

A notebook can be embedded using a standard iframe element. To create the embedding URL, copy the notebook URL (e.g. `https://your-sourcegraph-instance.com/notebooks/notebook-id`), and add the `/embed` prefix directly before the `/notebooks` segment:

  ```
  https://{your-sourcegraph-instance.com}/embed/notebooks/{notebook-id}
  ```

Once you have the embedding URL, create an iframe element and use the embedding URL as the `src` attribute value. See example iframe below:

```html
<iframe
  src="https://your-sourcegraph-instance.com/embed/notebooks/notebook-id"
  frameborder="0"
  sandbox="allow-scripts allow-same-origin allow-popups"
></iframe>
```

## Security
We recommend using the `sandbox` attribute to apply extra security restrictions to the iframe. Notebooks require three exceptions: `allow-scripts` allows executing Javascript scripts, `allow-same-origin` allows access to local storage and cookies, and `allow-popups` allows opening links in a separate tab.

## Enable embedding notebooks on private instances
Embedding is disabled by default on private instances. A site-admin can enable embedding by running the following GraphQL mutation in the API console, located at `https://{your-sourcegraph-instance.com}/api/console`:

```graphql
mutation {
  createFeatureFlag(name: "enable-embed-route", value: true) {
    ... on FeatureFlagBoolean {
      name
      value
    }
  }
}
```
