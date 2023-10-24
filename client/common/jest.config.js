// @ts-check

const config = require('../../jest.config.base')

const exportedConfig = {
  ...config,
  displayName: 'common',
  rootDir: __dirname,
  roots: ['<rootDir>/src'],
  verbose: true,
  setupFilesAfterEnv: [...(config.setupFilesAfterEnv || []), '<rootDir>/src/jestSetupAfterEnv.js'],
}

module.exports = exportedConfig
