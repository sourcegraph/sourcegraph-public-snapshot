// @ts-check

const customRules = {
  'forbid-class-name': require('./rules/forbid-class-name'),
  'forbid-link-href': require('./rules/forbid-link-href'),
}

module.exports = {
  rules: customRules,
  configs: {
    recommended: {
      plugins: ['react'],
      parserOptions: {
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
  },
}
