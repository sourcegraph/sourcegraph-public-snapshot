# Sourcegraph extension activation

Sourcegraph selectively activates each extension based on the `activationEvents` array in its `package.json`. This improves performance by only using the network and CPU for extensions that are needed.

There are 2 types of activation events:

- `["*"]`: always activate
- `["onLanguage:typescript"]`: activate for files of a language (multiple languages supported)

For simplicity, the extension creator sets `activationEvents` to `["*"]` but adjust this if your extension is language specific.

See the [list of languages](https://github.com/github/linguist/blob/master/lib/linguist/languages.yml), using the value for the `codemirror_mode` key.
