// @ts-check

const baseConfig = require('../../.eslintrc.js')

module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  // TODO(sqs): react/forbid-elements is to prep for usage outside of our monorepo
  rules: { 'no-console': 'off', 'react/forbid-elements': 'off' },
  overrides: baseConfig.overrides,
}
