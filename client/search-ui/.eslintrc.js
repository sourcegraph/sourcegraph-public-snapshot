// @ts-check

const baseConfig = require('../../.eslintrc.js')

module.exports = {
  extends: ['../../.eslintrc.js', ...baseConfig.extends],
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
}
