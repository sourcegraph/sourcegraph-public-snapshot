// @ts-check

const config = require('../../jest.config.base')

const exportedConfig = {
  ...config,
  displayName: 'client-api',
  rootDir: __dirname,
  roots: ['<rootDir>'],
  verbose: true,
}

module.exports = exportedConfig
