// @ts-check
const path = require('path')

const config = require('../../jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
const exportedConfig = {
  ...config,
  displayName: 'prerender',
  rootDir: __dirname,
  setupFiles: [
    ...config.setupFiles,
    path.join(__dirname, 'src', 'jestSetup.ts'),
  ]
}

module.exports = exportedConfig
