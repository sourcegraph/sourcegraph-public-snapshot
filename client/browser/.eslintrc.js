const baseConfig = require('../../.eslintrc')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json', __dirname + '/src/end-to-end/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
  rules: {
    // The browser extensions do not have the /help redirect
    '@sourcegraph/sourcegraph/forbid-docs-links': 'off',
  },
}
