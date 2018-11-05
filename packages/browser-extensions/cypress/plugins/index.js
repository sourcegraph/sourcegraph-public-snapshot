// ***********************************************************
// This example plugins/index.js can be used to load plugins
//
// You can change the location of this file or turn off loading
// the plugins file with the 'pluginsFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/plugins-guide
// ***********************************************************

const extensionLoader = require('cypress-browser-extension-plugin/loader')

module.exports = (on, config) => {
  on(
    'before:browser:launch',
    extensionLoader.load({
      source: 'build/chrome',
    })
  )
}
