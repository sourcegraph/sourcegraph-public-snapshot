// ------------------------------------
// #POSTCSS - LOAD CONFIG - INDEX
// ------------------------------------

'use strict'

var config = require('cosmiconfig')
var assign = require('object-assign')

var loadOptions = require('postcss-load-options/lib/options.js')
var loadPlugins = require('postcss-load-plugins/lib/plugins.js')

/**
 * @author Michael Ciniawsky (@michael-ciniawsky) <michael.ciniawsky@gmail.com>
 * @description Autoload Config for PostCSS
 *
 * @module postcss-load-config
 * @version 1.0.0
 *
 * @requires comsiconfig
 * @requires object-assign
 * @requires postcss-load-options
 * @requires postcss-load-plugins
 *
 * @method postcssrc
 *
 * @param  {Object} ctx Context
 * @param  {String} path Config Directory
 * @param  {Object} options Config Options
 *
 * @return {Promise} config  PostCSS Plugins, PostCSS Options
 */
module.exports = function postcssrc (ctx, path, options) {
  var defaults = { cwd: process.cwd(), env: process.env.NODE_ENV }

  ctx = assign(defaults, ctx)
  path = path || process.cwd()
  options = options || {}

  if (ctx.env === undefined) {
    process.env.NODE_ENV = 'development'
  }

  return config('postcss', options)
    .load(path)
    .then(function (result) {
      if (!result) {
        console.log(
          'PostCSS Config could not be loaded. Please check your PostCSS Config.'
        )
      }

      return result ? result.config : {}
    })
    .then(function (config) {
      if (typeof config === 'function') {
        config = config(ctx)
      } else {
        config = assign(config, ctx)
      }

      if (!config.plugins) {
        config.plugins = []
      }

      return {
        plugins: loadPlugins(config),
        options: loadOptions(config)
      }
    })
    .catch(function (err) {
      console.log(err)
    })
}
