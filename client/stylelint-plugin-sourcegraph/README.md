# stylelint-plugin-sourcegraph

Recommended Stylelint rules for the Sourcegraph repo.

## Setup

Update your `.stylelintrc.json` file to add the following configuration:

```js
  "plugins": [
    "@sourcegraph/stylelint-plugin-sourcegraph"
  ],
  "rules": {
    "filenames/match-regex": [2, "^.+\\.module(\\.scss)$"]
  }
```
