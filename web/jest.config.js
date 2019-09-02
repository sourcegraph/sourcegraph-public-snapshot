// @ts-check

/** @type {jest.InitialOptions} */
const config = require('../jest.config.base')

// Set testURL to match the value of SOURCEGRAPH_BASE_URL
if (process.env.SOURCEGRAPH_BASE_URL) {
  config.testURL = process.env.SOURCEGRAPH_BASE_URL
}

/** @type {jest.InitialOptions} */
module.exports = { ...config, displayName: 'web', rootDir: __dirname }
