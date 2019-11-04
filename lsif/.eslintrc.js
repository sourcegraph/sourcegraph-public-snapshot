module.exports = {
  extends: '../.eslintrc.js',
  rules: {
    'no-console': ['error'],
    'import/no-cycle': ['error'],
    'no-return-await': ['error'],
  },
  overrides: require('../.eslintrc').overrides,
}
