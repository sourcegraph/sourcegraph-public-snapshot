// @ts-check

/** @type {import('@jest/types').Config.InitialOptions} */
const config = require('../../../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
module.exports = {
  ...config,
  displayName: 'web-integration',
  rootDir: __dirname,
  testTimeout: 60000,
  preset: 'jest-puppeteer',
}
