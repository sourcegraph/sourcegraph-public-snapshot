# eslint-plugin-flow-vars

[![Build Status](https://travis-ci.org/zertosh/eslint-plugin-flow-vars.svg?branch=master)](https://travis-ci.org/zertosh/eslint-plugin-flow-vars)

An [eslint](https://github.com/eslint/eslint) plugin that makes flow type annotations global variables and marks declarations as used. Solves the problem of false positives with `no-undef` and `no-unused-vars` when using [`babel-eslint`](https://github.com/babel/babel-eslint).

## Usage

```sh
npm install eslint babel-eslint eslint-plugin-flow-vars
```

In your `.eslintrc`:

```json
{
  "parser": "babel-eslint",
  "plugins": [
    "flow-vars"
  ],
  "rules": {
    "flow-vars/define-flow-type": 1,
    "flow-vars/use-flow-type": 1
  }
}
```
