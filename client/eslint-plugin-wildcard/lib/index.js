// @ts-check

const customRules = {
  'forbid-class-name': require('./rules/forbid-class-name'),
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
