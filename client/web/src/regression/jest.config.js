// @ts-check

const config = require('../../../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
const exportedConfig = {
  ...config,
  displayName: 'web-integration',
  rootDir: __dirname,
  roots: ['<rootDir>'],
  testTimeout: 300000,
}

module.exports = exportedConfig
