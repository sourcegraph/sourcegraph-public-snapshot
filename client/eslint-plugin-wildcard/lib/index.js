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
          'error',
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
          'error',
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
