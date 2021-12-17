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
      rules: {
        'react/forbid-elements': [
          'off', // Currently disabled until first `<Button />` component is migrated: https://github.com/sourcegraph/codemod/issues/31
          {
            forbid: [
              {
                element: 'button',
                message: 'Use the <Button /> component from @sourcegraph/wildcard instead.',
              },
            ],
          },
        ],
        '@sourcegraph/wildcard/forbid-class-name': [
          'off', // Current disabled until first `<Badge />` component is migrated: https://github.com/sourcegraph/sourcegraph/issues/27622
          {
            forbid: [
              {
                className: 'badge',
                message: 'Use the <Badge /> component from @sourcegraph/wildcard instead.',
              },
            ],
          },
        ],
      },
    },
  },
}
