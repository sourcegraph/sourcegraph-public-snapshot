# Sourcegraph extension activation

To save resources, Sourcegraph will selectively activate an extension based on values in the `activationEvents` array in `package.json`.

There are three ways to set `activationEvents`:

1. Always activate: `["*"]`
1. Activate for a language: `["onLanguage:typescript"]`
1. Activate for multiple languages: `["onLanguage:typescript", "onLanguage:javascript"]`

For simplicity, the extension creator sets `activationEvents` to `["*"]` but adjust this if your extension is language specific.

See the [list of languages](https://github.com/github/linguist/blob/master/lib/linguist/languages.yml), using the value for the `codemirror_mode` key.

## Getting configuration upon activation

Extensions currently must wait until after `activate()` is called to get the configuration as it is not synchronously available. A temporary workaround is:

```typescript
export async function activate(): Promise<void> {
    // HACK: work around configuration not being synchronously available
    await new Promise(resolve => setTimeout(resolve, 100))

    // configuration can now be accessed
}```
