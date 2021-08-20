// @ts-check
const path = require('path')

const config = require('../../jest.config.base')

/** @type {jest.InitialOptions} */
const exportedConfig = {
  ...config,
  displayName: 'web',
  rootDir: __dirname,
  setupFiles: [
    ...config.setupFiles,
    'jest-canvas-mock', // mocking canvas is required for Monaco editor to work in unit tests
    path.join(__dirname, 'dev/mocks/mockEventLogger.ts'),
  ],
}

module.exports = exportedConfig
