'use strict'

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
                message: 'Use the Button component from @sourcegraph/wildcard',
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
                component: '<Badge />',
              },
            ],
          },
        ],
      },
    },
  },
}
