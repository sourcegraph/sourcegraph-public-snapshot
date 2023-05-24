const baseConfig = require('../../.eslintrc')

module.exports = {
  extends: '../../.eslintrc.js',
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
    {
      files: 'dev/**/*.ts',
      rules: {
        'no-console': 'off',
        'no-sync': 'off',
      },
    },
  ],
  rules: {
    'no-restricted-imports': [
      'error',
      {
        // Allow any imports from @sourcegraph/cody-* packages while those packages' APIs are being
        // stabilized.
        patterns: ['!@sourcegraph/cody-shared/*', '!@sourcegraph/cody-ui/*'],
      },
    ],
  },
}
