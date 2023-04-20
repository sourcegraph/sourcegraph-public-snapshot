// @ts-check

const baseConfig = require('../../.eslintrc.js')

module.exports = {
  root: true,
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
}
