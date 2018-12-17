// @ts-check

/** @type {jest.InitialOptions} */
const config = require('./jest.config.base')

/** @type {jest.InitialOptions} */
module.exports = {
  ...config,
  projects: ['client/browser/jest.config.js', 'shared/jest.config.js', 'web/jest.config.js'],
}
