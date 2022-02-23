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
  ],
}
