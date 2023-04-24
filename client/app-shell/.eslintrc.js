// @ts-check

const baseConfig = require('../../.eslintrc')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  overrides: [
    ...baseConfig.overrides,
    {
      files: 'dev/**/*.ts',
      rules: {
        'no-console': 'off',
        'no-sync': 'off',
      },
    },
  ],
  rules: {
    'react/react-in-jsx-scope': 'off',
    'react/jsx-filename-extension': [1, { extensions: ['.ts', '.tsx'] }],
    'id-length': 'off',
    'no-console': 'off',
    'no-void': ['error', { allowAsStatement: true }],
    '@typescript-eslint/no-floating-promises': ['error', { ignoreVoid: true }],
    'react/forbid-elements': 'off',
    'unicorn/filename-case': 'off',
  },
}
