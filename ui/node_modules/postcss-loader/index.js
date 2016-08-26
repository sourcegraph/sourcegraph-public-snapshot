var formatCodeFrame = require('babel-code-frame');
var loaderUtils     = require('loader-utils');
var postcss         = require('postcss');

function formatMessage(message, loc, source) {
    var formatted = message;
    if (loc) {
        formatted = formatted +
            ' (' + loc.line + ':' + loc.column + ')';
    }
    if (loc && source) {
        formatted = formatted +
            '\n\n' + formatCodeFrame(source, loc.line, loc.column) + '\n';
    }
    return formatted;
}

function PostCSSLoaderError(name, message, loc, source, error) {
    Error.call(this);
    Error.captureStackTrace(this, PostCSSLoaderError);
    this.name = name;
    this.error = error;
    this.message = formatMessage(message, loc, source);
    this.hideStack = true;
}

PostCSSLoaderError.prototype = Object.create(Error.prototype);
PostCSSLoaderError.prototype.constructor = PostCSSLoaderError;

module.exports = function (source, map) {
    if ( this.cacheable ) this.cacheable();

    var file   = this.resourcePath;
    var params = loaderUtils.parseQuery(this.query);

    var opts = {
        from: file,
        to:   file,
        map:  {
            inline:     params.sourceMap === 'inline',
            annotation: false
        }
    };

    if ( typeof map === 'string' ) map = JSON.parse(map);
    if ( map && map.mappings ) opts.map.prev = map;

    var options = this.options.postcss;
    if ( typeof options === 'function' ) {
        options = options.call(this, this);
    }

    var plugins;
    if ( typeof options === 'undefined' ) {
        plugins = [];
    } else if ( Array.isArray(options) ) {
        plugins = options;
    } else {
        plugins = options.plugins || options.defaults;
        opts.stringifier = options.stringifier;
        opts.parser      = options.parser;
        opts.syntax      = options.syntax;
    }
    if ( params.pack ) {
        plugins = options[params.pack];
        if ( !plugins ) {
            throw new Error('PostCSS plugin pack is not defined in options');
        }
    }

    if ( params.syntax ) {
        opts.syntax = require(params.syntax);
    }
    if ( params.parser ) {
        opts.parser = require(params.parser);
    }
    if ( params.stringifier ) {
        opts.stringifier = require(params.stringifier);
    }

    var loader   = this;
    var callback = this.async();

    if ( params.parser === 'postcss-js' ) {
        source = this.exec(source, this.resource);
    }

    // Allow plugins to add or remove postcss plugins
    plugins = this._compilation.applyPluginsWaterfall(
        'postcss-loader-before-processing',
        [].concat(plugins),
        params
    );

    postcss(plugins).process(source, opts)
        .then(function (result) {
            result.warnings().forEach(function (msg) {
                loader.emitWarning(msg.toString());
            });
            callback(null, result.css, result.map ? result.map.toJSON() : null);
        })
        .catch(function (error) {
            if ( error.name === 'CssSyntaxError' ) {
                callback(new PostCSSLoaderError(
                    'Syntax Error',
                    error.reason,
                    { line: error.line, column: error.column },
                    error.input.source));
            } else {
                callback(error);
            }
        });
};
