const baseConfig = require('../../.eslintrc')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: 'tsconfig.json',
  },
  overrides: baseConfig.overrides,
}
