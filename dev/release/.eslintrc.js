const baseConfig = require('../../.eslintrc')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: __dirname + '/tsconfig.json',
  },
  rules: { 'no-console': 'off' },
  overrides: baseConfig.overrides,
}
