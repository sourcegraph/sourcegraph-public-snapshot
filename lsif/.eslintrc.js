module.exports = {
  extends: '../.eslintrc.js',
  rules: {
    'no-console': ['error'],
    'import/no-cycle': ['error'],
    'no-return-await': ['error'],
    'no-shadow': ['error', { allow: ['ctx'] }],
  },
  overrides: require('../.eslintrc').overrides,
}
