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
                element: 'textarea',
                message: 'Use the <TextArea /> component from @sourcegraph/wildcard instead.',
              },
              {
                element: 'a',
                message: 'Use the <Link /> component from @sourcegraph/wildcard instead.',
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
        'no-restricted-imports': [
          'error',
          {
            paths: [
              {
                name: 'react-router-dom',
                importNames: ['Link'],
                message: 'Use the <Link /> component from @sourcegraph/wildcard instead.',
              },
            ],
          },
        ],
      },
    },
  },
}
