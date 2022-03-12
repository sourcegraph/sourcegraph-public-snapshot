// @ts-check

const config = require('../../../../jest.node.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
const exportedConfig = {
  ...config,
  displayName: 'browser-integration',
  rootDir: __dirname,
  verbose: true,
}

module.exports = exportedConfig
