// @ts-check
const baseConfig = require('../../.eslintrc.js')
module.exports = {
  extends: '../../.eslintrc.js',
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json'],
  },
  rules: {},
  overrides: [
    ...baseConfig.overrides,
    {
      files: ['src/*.test.*', 'src/testutils/**'],
      rules: {
        'import/extensions': ['error', 'never', { html: 'always' }],
      },
    },
  ],
}
