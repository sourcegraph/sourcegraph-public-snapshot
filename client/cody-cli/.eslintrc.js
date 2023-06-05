// @ts-check

const baseConfig = require('../../.eslintrc')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
  rules: {
    'id-length': 'off',
    'no-console': 'off',
    'no-restricted-imports': [
      'error',
      {
        patterns: ['!@sourcegraph/cody-shared/*'], // allow any imports from the @sourcegraph/cody-shared package
      },
    ],
    'unicorn/filename-case': 'off',
    'arrow-body-style': 'off',
    '@typescript-eslint/explicit-function-return-type': 'off',
  },
}
