// eslint-disable-next-line import/extensions
const baseConfig = require('../../.eslintrc.js')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json', __dirname + '/src/testing/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
  rules: {
    // We cannot use the versioned "/help" redirect in the browser extension
    '@sourcegraph/sourcegraph/forbid-docs-links': 'off',
  },
}
