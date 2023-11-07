// @ts-check

/** @type {import('@jest/types').Config.InitialOptions} */
const config = require('../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
module.exports = {
  ...config,
  displayName: 'browser',
  rootDir: __dirname,
  roots: ['<rootDir>/src'],
  modulePathIgnorePatterns: ['<rootDir>/.*runfiles.*', '.*/end-to-end/.*'], // TODO(sqs)
  setupFilesAfterEnv: [...(config.setupFilesAfterEnv || []), '<rootDir>/src/shared/jestSetupAfterEnv.js'],
}
