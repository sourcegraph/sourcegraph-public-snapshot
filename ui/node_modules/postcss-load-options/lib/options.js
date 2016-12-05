// ------------------------------------
// #POSTCSS - LOAD OPTIONS - OPTIONS
// ------------------------------------

'use strict'

/**
 *
 * @method options
 *
 * @param  {Object} options PostCSS Config
 *
 * @return {Object} options PostCSS Options
 */
module.exports = function options (options) {
  if (options.parser) {
    options.parser = require(options.parser)
  }

  if (options.syntax) {
    options.syntax = require(options.syntax)
  }

  if (options.stringifier) {
    options.stringifier = require(options.stringifier)
  }

  if (options.plugins) {
    delete options.plugins
  }

  return options
}
