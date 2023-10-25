// @ts-check
const config = require('../../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
const exportedConfig = {
  ...config,
  displayName: 'web-dev',
  rootDir: __dirname,
  roots: ['<rootDir>'],
}

module.exports = exportedConfig
