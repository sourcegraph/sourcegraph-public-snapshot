// @ts-check

const path = require('path')

const baseConfig = require('./jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
const config = {
  ...baseConfig,

  testEnvironment: __dirname + '/client/shared/dev/jest-node-environment.js',
  collectCoverage: false,
  roots: ['<rootDir>'],
  testTimeout: 300000,

  setupFiles: [require.resolve('ts-node/register/transpile-only')],
}

module.exports = config
