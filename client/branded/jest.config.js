// @ts-check

/** @type {import('@jest/types').Config.InitialOptions} */
const config = require('../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
module.exports = {
  ...config,
  displayName: 'branded',
  rootDir: __dirname,
}
