'use strict';

var utils = require('loader-utils');
var sass = require('node-sass');
var path = require('path');
var os = require('os');
var fs = require('fs');

// A typical sass error looks like this
var SassError = {
    message: 'invalid property name',
    column: 14,
    line: 1,
    file: 'stdin',
    status: 1
};
var resolveError = /Cannot resolve/;

/**
 * The sass-loader makes node-sass available to webpack modules.
 *
 * @param {string} content
 * @returns {*}
 */
module.exports = function (content) {
    var callback = this.async();
    var isSync = typeof callback !== 'function';
    var self = this;
    var resourcePath = this.resourcePath;
    var extensionMatcher = /\.(sass|scss|css)$/;
    var result;
    var fileExt;
    var opt;
    var contextMatch;
    var extension;

    /**
     * Enhances the sass error with additional information about what actually went wrong.
     *
     * @param {SassError} err
     */
    function formatSassError(err) {
        var msg = err.message;

        if (err.file === 'stdin') {
            err.file = resourcePath;
        }

        // The 'Current dir' hint of node-sass does not help us, we're providing
        // additional information by reading the err.file property
        msg = msg.replace(/\s*Current dir:\s*/, '');

        err.message = getFileExcerptIfPossible(err) +
            msg.charAt(0).toUpperCase() + msg.slice(1) + os.EOL +
            '      in ' + err.file + ' (line ' + err.line + ', column ' + err.column + ')';

        // Instruct webpack to hide the JS stack from the console
        // Usually you're only interested in the SASS stack in this case.
        err.hideStack = true;
    }

    /**
     * Returns an importer that uses webpack's resolving algorithm.
     *
     * It's important that the returned function has the correct number of arguments
     * (based on whether the call is sync or async) because otherwise node-sass doesn't exit.
     *
     * @returns {function}
     */
    function getWebpackImporter() {
        if (isSync) {
            return function syncWebpackImporter(url, context) {
                url = urlToRequest(url, context);
                context = normalizeContext(context);

                return syncResolve(self, url, context);
            };
        }
        return function asyncWebpackImporter(url, context, done) {
            url = urlToRequest(url, context);
            context = normalizeContext(context);

            asyncResolve(self, url, context, done);
        };
    }

    function urlToRequest(url, context) {
        contextMatch = context.match(extensionMatcher);

        // Add sass/scss/css extension if it is missing
        // The extension is inherited from importing resource or the default is used
        if (!url.match(extensionMatcher)) {
            extension = contextMatch && contextMatch[0] || fileExt;
            url = url + extension;
        }

        return utils.urlToRequest(url, opt.root);
    }

    function normalizeContext(context) {
        // The first file is 'stdin' when we're using the data option
        if (context === 'stdin') {
            context = resourcePath;
        }
        return path.dirname(context);
    }

    // When files have been imported via the includePaths-option, these files need to be
    // introduced to webpack in order to make them watchable.
    function addIncludedFilesToWebpack(includedFiles) {
        includedFiles.forEach(self.dependency);
    }

    this.cacheable();

    opt = utils.parseQuery(this.query);
    opt.data = content;

    // Skip empty files, otherwise it will stop webpack, see issue #21
    if (opt.data.trim() === '') {
        return callback(null, content);
    }

    // opt.outputStyle
    if (!opt.outputStyle && this.minimize) {
        opt.outputStyle = 'compressed';
    }

    // opt.sourceMap
    // Not using the `this.sourceMap` flag because css source maps are different
    // @see https://github.com/webpack/css-loader/pull/40
    if (opt.sourceMap) {
        // deliberately overriding the sourceMap option
        // this value is (currently) ignored by libsass when using the data input instead of file input
        // however, it is still necessary for correct relative paths in result.map.sources
        opt.sourceMap = this.options.output.path + '/sass.map';
    }

    // indentedSyntax is a boolean flag
    opt.indentedSyntax = Boolean(opt.indentedSyntax);
    fileExt = '.' + (opt.indentedSyntax? 'sass' : 'scss');

    // opt.importer
    opt.importer = getWebpackImporter();

    // start the actual rendering
    if (isSync) {
        try {
            result = sass.renderSync(opt);
            addIncludedFilesToWebpack(result.stats.includedFiles);
            return result.css.toString();
        } catch (err) {
            formatSassError(err);
            throw err;
        }
    }
    sass.render(opt, function onRender(err, result) {
        if (err) {
            formatSassError(err);
            callback(err);
            return;
        }

        if (result.map && result.map !== '{}') {
            result.map = JSON.parse(result.map);
            result.map.file = resourcePath;
            // The first source is 'stdin' according to libsass because we've used the data input
            // Now let's override that value with the correct relative path
            result.map.sources[0] = path.relative(self.options.output.path, resourcePath);
        } else {
            result.map = null;
        }

        addIncludedFilesToWebpack(result.stats.includedFiles);
        callback(null, result.css.toString(), result.map);
    });
};

/**
 * Tries to get an excerpt of the file where the error happened.
 * Uses err.line and err.column.
 *
 * Returns an empty string if the excerpt could not be retrieved.
 *
 * @param {SassError} err
 * @returns {string}
 */
function getFileExcerptIfPossible(err) {
    var content;

    try {
        content = fs.readFileSync(err.file, 'utf8');

        return os.EOL +
            content.split(os.EOL)[err.line - 1] + os.EOL +
            new Array(err.column - 1).join(' ') + '^' + os.EOL +
            '      ';
    } catch (err) {
        // If anything goes wrong here, we don't want any errors to be reported to the user
        return '';
    }
}

/**
 * Tries to resolve the given url synchronously. If a resolve error occurs, a second try for the same
 * module prefixed with an underscore is started.
 *
 * @param {object} loaderContext
 * @param {string} url
 * @param {string} context
 * @returns {object}
 */
function syncResolve(loaderContext, url, context) {
    var filename;
    var basename;

    try {
        filename = loaderContext.resolveSync(context, url);
    } catch (err) {
        basename = path.basename(url);
        if (requiresLookupForUnderscoreModule(err, basename)) {
            url = addUnderscoreToBasename(url, basename);
            return syncResolve(loaderContext, url, context);
        }

        // let the libsass do the rest job, e.g. search module in includePaths
        filename = path.join(path.dirname(url), removeUnderscoreFromBasename(basename));
    }

    return {
        file: filename
    };
}

/**
 * Tries to resolve the given url asynchronously. If a resolve error occurs, a second try for the same
 * module prefixed with an underscore is started.
 *
 * @param {object} loaderContext
 * @param {string} url
 * @param {string} context
 * @param {function} done
 */
function asyncResolve(loaderContext, url, context, done) {
    loaderContext.resolve(context, url, function onWebpackResolve(err, filename) {
        var basename;

        if (err) {
            basename = path.basename(url);
            if (requiresLookupForUnderscoreModule(err, basename)) {
                url = addUnderscoreToBasename(url, basename);
                return asyncResolve(loaderContext, url, context, done);
            }

            // Let libsass do the rest of the job, like searching for the module in includePaths
            filename = path.join(path.dirname(url), removeUnderscoreFromBasename(basename));
        }

        // Use self.loadModule() before calling done() to make imported files available to
        // other webpack tools like postLoaders etc.?

        done({
            file: filename
        });
    });
}

/**
 * Check whether its a resolve error and the basename does *not* start with an underscore.
 *
 * @param {Error} err
 * @param {string} basename
 * @returns {boolean}
 */
function requiresLookupForUnderscoreModule(err, basename) {
    return resolveError.test(err.message) && basename.charAt(0) !== '_';
}

/**
 * @param {string} url
 * @param {string} basename
 * @returns {string}
 */
function addUnderscoreToBasename(url, basename) {
    return url.slice(0, -basename.length) + '_' + basename;
}

/**
 * @param {string} basename
 * @returns {string}
 */
function removeUnderscoreFromBasename(basename) {
  return basename[0] === '_' ? basename.substring(1) : basename;
}
