# Sourcegraph extension activation

Sourcegraph selectively activates each extension based on the `activationEvents` array in its `package.json`. This improves performance by only using the network and CPU for extensions when necessary.

There are two types of activation events:

- `["*"]`: always activate
- `["onLanguage:typescript"]`: activate for files of a language (multiple languages supported)

For simplicity, the extension creator sets `activationEvents` to `["*"]`. Adjust this if your extension is language-specific.

## Determining the correct language value

Search this [list of languages](https://github.com/github/linguist/blob/master/lib/linguist/languages.yml) to find the value assigned to the `codemirror_mode` key for that language.
