// @ts-check

/** @type {import('@jest/types').Config.InitialOptions} */
const config = require('../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
module.exports = {
  ...config,
  displayName: 'cody',
  rootDir: __dirname,
  testPathIgnorePatterns: ['integration-test'],
}
