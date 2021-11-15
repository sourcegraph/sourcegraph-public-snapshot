module.exports = {
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
  },
}
