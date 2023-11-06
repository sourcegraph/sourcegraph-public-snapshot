// @ts-check

const config = require('../../jest.config.base')

const exportedConfig = {
  ...config,
  displayName: 'build-config',
  rootDir: __dirname,
  roots: ['<rootDir>'],
  verbose: true,
}

module.exports = exportedConfig
