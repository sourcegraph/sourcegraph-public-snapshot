// @ts-check

const baseConfig = require('../../.eslintrc.js')

module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  rules: {
    'no-console': 'off',
    'jsdoc/check-indentation': 'off',
  },
  overrides: baseConfig.overrides,
}
