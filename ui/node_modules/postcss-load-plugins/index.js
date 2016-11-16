// ------------------------------------
// #POSTCSS - LOAD PLUGINS - INDEX
// ------------------------------------

'use strict'

var config = require('cosmiconfig')
var assign = require('object-assign')

var loadPlugins = require('./lib/plugins')

/**
 * @author Michael Ciniawsky (@michael-ciniawsky) <michael.ciniawsky@gmail.com>
 * @description Autoload Plugins for PostCSS
 *
 * @module postcss-load-plugins
 * @version 1.0.0
 *
 * @requires cosmiconfig
 * @requires object-assign
 * @requires ./lib/plugins.js
 *
 * @method pluginsrc
 *
 * @param  {Object} ctx Context
 * @param  {String} path Directory
 * @param  {Object} options Options
 *
 * @return {Array} config PostCSS Plugins
 */
module.exports = function pluginsrc (ctx, path, options) {
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
      if (result === undefined) {
        console.log(
          'PostCSS Plugins could not be loaded. Please check your PostCSS Config.'
        )
      }

      return result ? result.config : {}
    })
    .then(function (plugins) {
      if (typeof plugins === 'function') {
        plugins = plugins(ctx)
      }
      if (typeof result === 'object') {
        plugins = assign(plugins, ctx)
      }

      if (!plugins.plugins) {
        plugins.plugins = []
      }

      return loadPlugins(plugins)
    })
    .catch(function (err) {
      console.log(err)
    })
}
