// ------------------------------------
// #POSTCSS - LOAD PLUGINS - PLUGINS
// ------------------------------------

'use strict'

/**
 * @method plugins
 *
 * @param {Object} config PostCSS Config
 *
 * @return {Array} plugins PostCSS Plugins
 */
module.exports = function plugins (config) {
  var plugins = []

  if (Array.isArray(config.plugins)) {
    plugins = config.plugins

    if (plugins.length > 0) {
      plugins.forEach(function (plugin) {
        if (typeof plugin !== 'function') {
          throw new TypeError(
            plugin + ' must be a function, did you require() it ?'
          )
        }
      })
    }

    return plugins
  } else {
    config = config.plugins

    var load = function (plugin, options) {
      if (options === null || Object.keys(options).length === 0) {
        try {
          return require(plugin)
        } catch (err) {
          console.log(err)
        }
      } else {
        try {
          return require(plugin)(options)
        } catch (err) {
          console.log(err)
        }
      }
    }

    Object.keys(config)
      .filter(function (plugin) {
        return config[plugin] !== false ? plugin : ''
      })
      .forEach(function (plugin) {
        plugin = load(plugin, config[plugin])

        if (plugin.default && typeof plugin.default === 'function') {
          plugin = plugin.default
        }

        return plugins.push(plugin)
      })

    return plugins
  }
}
