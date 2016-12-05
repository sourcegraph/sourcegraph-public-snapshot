// ------------------------------------
// #POSTCSS - LOAD OPTIONS
// ------------------------------------

'use strict'

var config = require('cosmiconfig')
var assign = require('object-assign')

var loadOptions = require('./lib/options')

/**
 * @author Michael Ciniawsky (@michael-ciniawsky) <michael.ciniawsky@gmail.com>
 * @description Autoload Options for PostCSS
 *
 *
 * @module postcss-load-options
 * @version 1.0.0
 *
 * @requires cosmiconfig
 * @requires object-assign
 * @requires lib/options
 *
 * @method optionsrc
 *
 * @param  {Object} ctx Context
 * @param  {String} path Directory
 * @param  {Object} options Options
 * @return {Object} options PostCSS Options
 */
module.exports = function optionsrc (ctx, path, options) {
  var defaults = { cwd: process.cwd(), env: process.env.NODE_ENV }

  ctx = assign(defaults, ctx) || defaults
  path = path || process.cwd()
  options = options || {}

  return config('postcss', options)
    .load(path)
    .then(function (result) {
      if (result === undefined) {
        console.log(
          'PostCSS Options could not be loaded. Please check your PostCSS Config.'
        )
      }
      result = result === undefined ? { config: {} } : result
      return result
    })
    .then(function (options) {
      if (typeof options === 'function') {
        options = options(ctx)
      }

      if (typeof options === 'object') {
        options = assign(options, ctx)
      }

      return options
    })
    .then(function (options) {
      return loadOptions(options)
    })
    .catch(function (err) {
      console.log(err)
    })
}
