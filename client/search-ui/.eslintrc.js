// @ts-check

const baseConfig = require('../../.eslintrc.js')

module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
  rules: {
    // We cannot use the versioned "/help" redirect in the VS Code extension
    '@sourcegraph/sourcegraph/forbid-docs-links': 'off',
  },
}
