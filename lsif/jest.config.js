// @ts-check

/** @type {jest.InitialOptions} */
const config = require('../jest.config.base')

/** @type {jest.InitialOptions} */
module.exports = { ...config, setupFilesAfterEnv: ['./jest.setup.js'], displayName: 'lsif', rootDir: __dirname }
