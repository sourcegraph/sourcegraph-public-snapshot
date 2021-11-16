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
      files: ['src/stores/global.ts', 'src/__mocks__/zustand.ts'],
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
            message: 'Avoid creating multiple global stores. Use the existing store in client/web/src/stores/global.ts',
          },
        ],
      },
    ],
  },
}
