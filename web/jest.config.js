// @ts-check

/** @type {jest.InitialOptions} */
const config = require('../jest.config.base')

/** @type {jest.InitialOptions} */
module.exports = {
  ...config,
  displayName: 'web',
  rootDir: __dirname,

  // Set testURL to match the value of SOURCEGRAPH_BASE_URL
  testURL: process.env.SOURCEGRAPH_BASE_URL || undefined,
}
