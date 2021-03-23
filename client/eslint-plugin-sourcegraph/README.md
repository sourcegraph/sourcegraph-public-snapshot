# eslint-plugin-sourcegraph

Custom ESLint rules for Sourcegraph. This package should only be used within
the main Sourcegraph project, and isn't intended for reuse by other packages in
the Sourcegraph organisation.

## Rules

Rules are defined in `lib/rules`. At present, one rule is available.

### `check-help-links`

This rule parses `Link` and `a` elements in JSX/TSX files. If a list of valid
docsite pages is provided, elements that point to a `/help/*` link are checked
against that list: if they don't exist, a linting error is raised.

The list of docsite pages is provided either via the `DOCSITE_LIST` environment
variable, which should be a newline separated list of pages as outputted by
`docsite ls`, or via the `docsiteList` rule option, which is the same data as
an array.

If neither of these are set, then the rule will silently succeed.

## Testing

Unit tests can be run with:

```sh
yarn test
```
