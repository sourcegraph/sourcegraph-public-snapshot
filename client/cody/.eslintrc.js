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
    'react/react-in-jsx-scope': 'off',
    'react/jsx-filename-extension': [1, { extensions: ['.ts', '.tsx'] }],
    'id-length': 'off',
    'no-console': 'off',
    'no-restricted-imports': ['error', { paths: ['!highlight.js'] }],
    'react/forbid-elements': 'off',
    'unicorn/filename-case': 'off',
  },
}
