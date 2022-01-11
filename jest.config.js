// @ts-check

/** @type {import('@jest/types').Config.InitialOptions} */
const config = require('./jest.config.base')

/** @type {import('@jest/types').Config.InitialOptions} */
module.exports = {
  projects: [
    'client/browser/jest.config.js',
    'client/build-config/jest.config.js',
    'client/common/jest.config.js',
    'client/codeintellify/jest.config.js',
    'client/shared/jest.config.js',
    'client/branded/jest.config.js',
    'client/web/jest.config.js',
    'client/wildcard/jest.config.js',
    'client/template-parser/jest.config.js',
    'client/storybook/jest.config.js',
  ],
}
