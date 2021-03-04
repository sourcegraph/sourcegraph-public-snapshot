// @ts-check

/** @type {jest.InitialOptions} */
const config = require('./jest.config.base')

/** @type {jest.InitialOptions} */
module.exports = {
  projects: [
    'client/extension-browser/jest.config.js',
    'client/ui-kit-legacy-shared/jest.config.js',
    'client/ui-kit-legacy-branded/jest.config.js',
    'client/app-web/jest.config.js',
    'client/ui-kit/.storybook/jest.config.js',
  ],
}
