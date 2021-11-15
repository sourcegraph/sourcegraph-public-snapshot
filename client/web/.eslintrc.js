const baseConfig = require('../../.eslintrc')
module.exports = {
  extends: ['../../.eslintrc.js', '@sourcegraph/wildcard'],
  parserOptions: {
    ...baseConfig.parserOptions,
    project: [__dirname + '/tsconfig.json', __dirname + '/src/**/tsconfig.json'],
  },
  overrides: baseConfig.overrides,
}
