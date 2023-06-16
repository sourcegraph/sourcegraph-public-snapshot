// @ts-check

const baseConfig = require('../../.eslintrc.js')

module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
  rules: {
    '@typescript-eslint/no-require-imports': 'off',
    '@typescript-eslint/no-var-requires': 'off',
    '@typescript-eslint/unbound-method': 'off',
    'no-restricted-imports': 'off',
    'react/forbid-dom-props': 'off',
    'import/no-default-export': 'off',
    'no-console': 'off',
    'no-duplicate-imports': 'off',
    'arrow-body-style': 'off',
    '@typescript-eslint/explicit-function-return-type': 'off',
    '@typescript-eslint/consistent-type-definitions': 'off',
    'ban/ban': 'off',
    'no-sync': 'off',
  },
}
