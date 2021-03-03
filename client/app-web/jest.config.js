// @ts-check

const config = require('../../jest.config.base')

/** @type {jest.InitialOptions} */
const exportedConfig = {
  ...config,
  displayName: 'web',
  rootDir: __dirname,
  setupFiles: [...config.setupFiles, 'jest-canvas-mock'], // mocking canvas is required for Monaco editor to work in unit tests
}

module.exports = exportedConfig
