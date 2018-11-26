# Sourcegraph extension manifest file - package.json

Sourcegraph extensions use a `package.json` file for metadata and configuration.

## Fields

Name | Required | Type | Details
---- |:--------:| ---- | -------
`name` | ✔️ | `string` | Extension identifier: all lowercase, alphanumeric with hyphens and underscores.
`title` | ✔️ | `string`| The name displayed in the extension registry. Can be used to indicate a [work-in-progress extension](publishing.md#wip-extensions).
`description` | ✔️ | `string` | A short description of what your extension is and does.
`version` | | `string` | [Semantic versioning](https://semver.org/) format.
`publisher` | ✔️ | `string` | Your [Sourcegraph username](development_environment#sourcegraph-com-account-and-the-sourcegraph-cli)
`license` | | `string` | The type of license chosen.
`main` | | `string` | Path to the transpiled JavaScript file for your extension.
`contributes` | | `object` | An object describing the contributions (extension points) for this extension (e.g. menus, buttons, configuration and more).
[`activationEvents`](activation.md) | Y | `array` | An array of event strings for activating your extension.
`devDependencies` | | `object` | npm dependencies needed for development.
`dependencies` | | `object` | npm dependencies needed at runtime.
`scripts` | ✔️ | `object` | npm's scripts with Sourcegraph specific entries such as `sourcegraph:prepublish`.
`browserslist` | | `string` | Modern list of browsers for build tools to target when transpiling.
`repository` | | `object` | npm field for the repository location.

See the [npm package.json](https://docs.npmjs.com/creating-a-package-json-file) documentation for other fields.

**Note:** Including the `repository` field is recommended so anyone can follow the link from the extension detail page and view the source code.

```json
"repository": {
  "type": "git",
  "url": "https://github.com/sourcegraph/sourcegraph-codecov.git"
}
```

## Example

Here is an example `package.json` created by the [Sourcegraph extension creator](https://docs.sourcegraph.com/extensions/authoring/creating#creating-an-extension-the-easy-way).

```json
{
  "name": "my-extension",
  "title": "WIP: My Extension",
  "description": "An awesome Sourcegraph extension",
  "publisher": "your-sourcegraph-username",
  "activationEvents": [
    "*"
  ],
  "contributes": {
    "actions": [
      {}
    ],
    "menus": {
      "editor/title": [],
      "commandPalette": []
    },
    "configuration": {}
  },
  "version": "0.0.0-DEVELOPMENT",
  "license": "MIT",
  "main": "dist/my-extension.js",
  "scripts": {
    "tslint": "tslint -p tsconfig.json './src/**/*.ts'",
    "typecheck": "tsc -p tsconfig.json",
    "build": "parcel build --out-file dist/my-extension.js src/my-extension.ts",
    "serve": "parcel serve --no-hmr --out-file dist/my-extension.js src/my-extension.ts",
    "watch:typecheck": "tsc -p tsconfig.json -w",
    "watch:build": "tsc -p tsconfig.dist.json -w",
    "sourcegraph:prepublish": "npm run build"
  },
  "browserslist": [
    "last 1 Chrome versions",
    "last 1 Firefox versions",
    "last 1 Edge versions",
    "last 1 Safari versions"
  ],
  "devDependencies": {
    "@sourcegraph/tsconfig": "^3.0.0",
    "@sourcegraph/tslint-config": "^12.0.0",
    "parcel-bundler": "^1.10.3",
    "sourcegraph": "^19.0.3",
    "tslint": "^5.11.0",
    "typescript": "^3.1.6"
  }
}
```
