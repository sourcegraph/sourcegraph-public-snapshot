var loaderUtils = require('loader-utils');
var postcss     = require('postcss');

module.exports = function (source, map) {
    if ( this.cacheable ) this.cacheable();

    var file   = this.resourcePath;
    var params = loaderUtils.parseQuery(this.query);

    var opts = {
        from: file,
        to:   file,
        map:  {
            inline:     false,
            annotation: false
        }
    };

    if ( typeof map === 'string' ) map = JSON.parse(map);
    if ( map && map.mappings ) opts.map.prev = map;

    if ( params.syntax )      opts.syntax      = require(params.syntax);
    if ( params.parser )      opts.parser      = require(params.parser);
    if ( params.stringifier ) opts.stringifier = require(params.stringifier);

    var plugins = this.options.postcss;
    if ( typeof plugins === 'function' ) {
        plugins = plugins.call(this, this);
    }

    if ( typeof plugins === 'undefined' ) {
        plugins = [];
    } else if ( params.pack ) {
        plugins = plugins[params.pack];
    } else if ( !Array.isArray(plugins) ) {
        plugins = plugins.defaults;
    }

    var loader   = this;
    var callback = this.async();

    if ( params.parser === 'postcss-js' ) {
        source = this.exec(source, this.resource);
    }

    postcss(plugins)
        .process(source, opts).then(function (result) {
            result.warnings().forEach(function (msg) {
                loader.emitWarning(msg.toString());
            });
            callback(null, result.css, result.map);
        })
        .catch(function (error) {
            if ( error.name === 'CssSyntaxError' ) {
                loader.emitError(error.message + error.showSourceCode());
                callback();
            } else {
                callback(error);
            }
        });
};
