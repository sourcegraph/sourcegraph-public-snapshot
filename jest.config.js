// @ts-check

/** @type {jest.InitialOptions} */
const config = require('./jest.config.base')

/** @type {jest.InitialOptions} */
module.exports = {
  projects: [
    'client/browser/jest.config.js',
    'client/shared/jest.config.js',
    'client/branded/jest.config.js',
    'client/web/jest.config.js',
    'client/wildcard/jest.config.js',
    'client/storybook/jest.config.js',
  ],
}
