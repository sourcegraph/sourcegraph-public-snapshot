const baseConfig = require('../../.eslintrc')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: 'tsconfig.json',
  },
  rules: {
    'rxjs/no-async-subscribe': 'off', // https://github.com/cartant/eslint-plugin-rxjs/issues/46

    'no-console': ['error'],
    'import/no-cycle': ['error'],
    'no-return-await': ['error'],
    'no-shadow': ['error', { allow: ['ctx'] }],
  },
  overrides: baseConfig.overrides,
}
