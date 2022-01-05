const baseConfig = require('../../.eslintrc')
module.exports = {
  extends: ['../../.eslintrc.js', 'plugin:@sourcegraph/wildcard/recommended'],
  plugins: ['@sourcegraph/wildcard'],
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json', __dirname + '/src/**/tsconfig.json'],
  },
  overrides: [
    ...baseConfig.overrides,
    {
      files: ['src/stores/**.ts', 'src/__mocks__/zustand.ts'],
      rules: { 'no-restricted-imports': 'off' },
    },
  ],
  rules: {
    'no-restricted-imports': [
      'error',
      {
        paths: [
          {
            name: 'zustand',
            importNames: ['default'],
            message:
              'Our Zustand stores should be created in a single place. Create this store in client/web/src/stores',
          },
          {
            name: '@sourcegraph/wildcard',
            importNames: ['Tooltip'],
            message:
              'Please ensure there is only a single `<Tooltip />` component present in the React tree. To display a specific tooltip, you can add the `data-tooltip` attribute to the relevant element.',
          },
        ],
      },
    ],
  },
}
