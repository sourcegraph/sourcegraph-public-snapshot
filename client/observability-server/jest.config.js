// @ts-check

const config = require('../../jest.config.base')

const exportedConfig = {
  ...config,
  displayName: 'observability',
  rootDir: __dirname,
  roots: ['<rootDir>'],
  verbose: true,
}

module.exports = exportedConfig
