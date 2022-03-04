const baseConfig = require('../../.eslintrc.js')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json', __dirname + '/src/testing/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
  rules: {
    // The browser extensions do not have the /help redirect
    '@sourcegraph/sourcegraph/forbid-docs-links': 'off',
  },
}
