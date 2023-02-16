// @ts-check
const path = require('path')

const config = require('../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
const exportedConfig = {
  ...config,
  displayName: 'backstage-common',
  rootDir: __dirname,
  setupFiles: [...config.setupFiles],
}

module.exports = exportedConfig
