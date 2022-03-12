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

  setupFiles: [
    require.resolve('ts-node/register/transpile-only'),
    // Needed for reusing API functions that use fetch
    // Neither NodeJS nor JSDOM have fetch + AbortController yet
    require.resolve('abort-controller/polyfill'),
    path.join(__dirname, 'client/shared/dev/fetch'),
  ],
}

module.exports = config
