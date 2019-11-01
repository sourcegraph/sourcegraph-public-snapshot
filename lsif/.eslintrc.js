module.exports = {
  extends: '../.eslintrc.js',
  rules: {
    'no-console': ['error'],
    'import/no-cycle': ['error'],
  },
  overrides: require('../.eslintrc').overrides,
}
