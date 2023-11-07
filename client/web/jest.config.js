// @ts-check
const path = require('path')

const config = require('../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
const exportedConfig = {
  ...config,
  displayName: 'web',
  rootDir: __dirname,
  setupFiles: [...config.setupFiles, path.join(__dirname, '../shared/dev/mockEventLogger.ts')],
}

module.exports = exportedConfig
