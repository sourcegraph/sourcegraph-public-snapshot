// @ts-check

/** @type {jest.InitialOptions} */
const config = require('../jest.config.base')

/** @type {jest.InitialOptions} */
const exportedConfig = { ...config, displayName: 'web', rootDir: __dirname }

if (process.env.SOURCEGRAPH_BASE_URL) {
  exportedConfig.testURL = process.env.SOURCEGRAPH_BASE_URL
  // exportedConfig.testEnvironment = 'node'
}

/** @type {jest.InitialOptions} */
module.exports = exportedConfig
