'use strict';

/**
 * Copied from https://github.com/jrburke/r.js/blob/master/build/jslib/parse.js with a couple
 * of changes to make it run in node.
 */

/**
 * @license Copyright (c) 2010-2011, The Dojo Foundation All Rights Reserved.
 * Available via the MIT or new BSD license.
 * see: http://github.com/jrburke/requirejs for details
 */

/*jshint plusplus: false, strict: false, maxdepth: 6, maxcomplexity: 21, maxstatements: 28 */
/*global define: false */

var uglify = require('uglify-js'),
	parser = uglify.parser,
    processor = uglify.uglify,
    ostring = Object.prototype.toString,
    isArray;

if (Array.isArray) {
    isArray = Array.isArray;
} else {
    isArray = function (it) {
        return ostring.call(it) === "[object Array]";
    };
}

/**
 * Determines if the AST node is an array literal
 */
function isArrayLiteral(node) {
    return node[0] === 'array';
}

/**
 * Determines if the AST node is an object literal
 */
function isObjectLiteral(node) {
    return node[0] === 'object';
}

/**
 * Converts a regular JS array of strings to an AST node that
 * represents that array.
 * @param {Array} ary
 * @param {Node} an AST node that represents an array of strings.
 */
function toAstArray(ary) {
    var output = [
        'array',
        []
    ],
    i, item;

    for (i = 0; (item = ary[i]); i++) {
        output[1].push([
            'string',
            item
        ]);
    }

    return output;
}

/**
 * Validates a node as being an object literal (like for i18n bundles)
 * or an array literal with just string members. If an array literal,
 * only return array members that are full strings. So the caller of
 * this function should use the return value as the new value for the
 * node.
 *
 * This function does not need to worry about comments, they are not
 * present in this AST.
 *
 * @param {Node} node an AST node.
 *
 * @returns {Node} an AST node to use for the valid dependencies.
 * If null is returned, then it means the input node was not a valid
 * dependency.
 */
function validateDeps(node) {
    var newDeps = ['array', []],
        arrayArgs, i, dep;

    if (!node) {
        return null;
    }

    if (isObjectLiteral(node) || node[0] === 'function') {
        return node;
    }

    //Dependencies can be an object literal or an array.
    if (!isArrayLiteral(node)) {
        return null;
    }

    arrayArgs = node[1];

    for (i = 0; i < arrayArgs.length; i++) {
        dep = arrayArgs[i];
        if (dep[0] === 'string') {
            newDeps[1].push(dep);
        }
    }
    return newDeps[1].length ? newDeps : null;
}

/**
 * Gets dependencies from a node, but only if it is an array literal,
 * and only if the dependency is a string literal.
 *
 * This function does not need to worry about comments, they are not
 * present in this AST.
 *
 * @param {Node} node an AST node.
 *
 * @returns {Array} of valid dependencies.
 * If null is returned, then it means the input node was not a valid
 * array literal, or did not have any string literals..
 */
function getValidDeps(node) {
    var newDeps = [],
        arrayArgs, i, dep;

    if (!node) {
        return null;
    }

    if (isObjectLiteral(node) || node[0] === 'function') {
        return null;
    }

    //Dependencies can be an object literal or an array.
    if (!isArrayLiteral(node)) {
        return null;
    }

    arrayArgs = node[1];

    for (i = 0; i < arrayArgs.length; i++) {
        dep = arrayArgs[i];
        if (dep[0] === 'string') {
            newDeps.push(dep[1]);
        }
    }
    return newDeps.length ? newDeps : null;
}

/**
 * Main parse function. Returns a string of any valid require or define/require.def
 * calls as part of one JavaScript source string.
 * @param {String} moduleName the module name that represents this file.
 * It is used to create a default define if there is not one already for the file.
 * This allows properly tracing dependencies for builds. Otherwise, if
 * the file just has a require() call, the file dependencies will not be
 * properly reflected: the file will come before its dependencies.
 * @param {String} moduleName
 * @param {String} fileName
 * @param {String} fileContents
 * @param {Object} options optional options. insertNeedsDefine: true will
 * add calls to require.needsDefine() if appropriate.
 * @returns {String} JS source string or null, if no require or define/require.def
 * calls are found.
 */
function parse(moduleName, fileName, fileContents, options) {
    options = options || {};

    //Set up source input
    var moduleDeps = [],
        result = '',
        moduleList = [],
        needsDefine = true,
        astRoot = parser.parse(fileContents),
        i, moduleCall, depString;

    parse.recurse(astRoot, function (callName, config, name, deps) {
        //If name is an array, it means it is an anonymous module,
        //so adjust args appropriately. An anonymous module could
        //have a FUNCTION as the name type, but just ignore those
        //since we just want to find dependencies.
        if (name && isArrayLiteral(name)) {
            deps = name;
            name = null;
        }

        if (!(deps = getValidDeps(deps))) {
            deps = [];
        }

        //Get the name as a string literal, if it is available.
        if (name && name[0] === 'string') {
            name = name[1];
        } else {
            name = null;
        }

        if (callName === 'define' && (!name || name === moduleName)) {
            needsDefine = false;
        }

        if (!name) {
            //If there is no module name, the dependencies are for
            //this file/default module name.
            moduleDeps = moduleDeps.concat(deps);
        } else {
            moduleList.push({
                name: name,
                deps: deps
            });
        }

        //If define was found, no need to dive deeper, unless
        //the config explicitly wants to dig deeper.
        return !options.findNestedDependencies;
    }, options);

    if (options.insertNeedsDefine && needsDefine) {
        result += 'require.needsDefine("' + moduleName + '");';
    }

    if (moduleDeps.length || moduleList.length) {
        for (i = 0; (moduleCall = moduleList[i]); i++) {
            if (result) {
                result += '\n';
            }

            //If this is the main module for this file, combine any
            //"anonymous" dependencies (could come from a nested require
            //call) with this module.
            if (moduleCall.name === moduleName) {
                moduleCall.deps = moduleCall.deps.concat(moduleDeps);
                moduleDeps = [];
            }

            depString = moduleCall.deps.length ? '["' + moduleCall.deps.join('","') + '"]' : '[]';
            result += 'define("' + moduleCall.name + '",' + depString + ');';
        }
        if (moduleDeps.length) {
            if (result) {
                result += '\n';
            }
            depString = moduleDeps.length ? '["' + moduleDeps.join('","') + '"]' : '[]';
            result += 'define("' + moduleName + '",' + depString + ');';
        }
    }

    return result ? result : null;
}

//Add some private methods to object for use in derived objects.
parse.isArray = isArray;
parse.isObjectLiteral = isObjectLiteral;
parse.isArrayLiteral = isArrayLiteral;

/**
 * Handles parsing a file recursively for require calls.
 * @param {Array} parentNode the AST node to start with.
 * @param {Function} onMatch function to call on a parse match.
 * @param {Object} [options] This is normally the build config options if
 * it is passed.
 * @param {Function} [recurseCallback] function to call on each valid
 * node, defaults to parse.parseNode.
 */
parse.recurse = function (parentNode, onMatch, options, recurseCallback) {
    var hasHas = options && options.has,
        i, node;

    recurseCallback = recurseCallback || this.parseNode;

    if (isArray(parentNode)) {
        for (i = 0; i < parentNode.length; i++) {
            node = parentNode[i];
            if (isArray(node)) {
                //If has config is in play, if calls have been converted
                //by this point to be true/false values. So, if
                //options has a 'has' value, skip if branches that have
                //literal false values.

                //uglify returns if constructs in an array:
                //[0]: 'if'
                //[1]: the condition, ['name', true | false] for the has replaced case.
                //[2]: the block to process if true
                //[3]: the block to process if false
                //For if/else if/else, the else if is in the [3],
                //so only ever have to deal with this structure.
                if (hasHas && node[0] === 'if' && node[1] && node[1][0] === 'name' &&
                    (node[1][1] === 'true' || node[1][1] === 'false')) {
                    if (node[1][1] === 'true') {
                        this.recurse([node[2]], onMatch, options, recurseCallback);
                    } else {
                        this.recurse([node[3]], onMatch, options, recurseCallback);
                    }
                } else {
                    if (recurseCallback(node, onMatch)) {
                        //The onMatch indicated parsing should
                        //stop for children of this node.
                        continue;
                    }
                    this.recurse(node, onMatch, options, recurseCallback);
                }
            }
        }
    }
};

/**
 * Determines if the file defines require().
 * @param {String} fileName
 * @param {String} fileContents
 * @returns {Boolean}
 */
parse.definesRequire = function (fileName, fileContents) {
    var astRoot = parser.parse(fileContents);
    return this.nodeHasRequire(astRoot);
};

/**
 * Finds require("") calls inside a CommonJS anonymous module wrapped in a
 * define(function(require, exports, module){}) wrapper. These dependencies
 * will be added to a modified define() call that lists the dependencies
 * on the outside of the function.
 * @param {String} fileName
 * @param {String} fileContents
 * @returns {Array} an array of module names that are dependencies. Always
 * returns an array, but could be of length zero.
 */
parse.getAnonDeps = function (fileName, fileContents) {
    var astRoot = parser.parse(fileContents),
        defFunc = this.findAnonDefineFactory(astRoot);

    return parse.getAnonDepsFromNode(defFunc);
};

/**
 * Finds require("") calls inside a CommonJS anonymous module wrapped
 * in a define function, given an AST node for the definition function.
 * @param {Node} node the AST node for the definition function.
 * @returns {Array} and array of dependency names. Can be of zero length.
 */
parse.getAnonDepsFromNode = function (node) {
    var deps = [],
        funcArgLength;

    if (node) {
        this.findRequireDepNames(node, deps);

        //If no deps, still add the standard CommonJS require, exports, module,
        //in that order, to the deps, but only if specified as function args.
        //In particular, if exports is used, it is favored over the return
        //value of the function, so only add it if asked.
        funcArgLength = node[2] && node[2].length;
        if (funcArgLength) {
            deps = (funcArgLength > 1 ? ["require", "exports", "module"] :
                    ["require"]).concat(deps);
        }
    }
    return deps;
};

/**
 * Finds the function in define(function (require, exports, module){});
 * @param {Array} node
 * @returns {Boolean}
 */
parse.findAnonDefineFactory = function (node) {
    var callback, i, n, call, args;

    if (isArray(node)) {
        if (node[0] === 'call') {
            call = node[1];
            args = node[2];
            if ((call[0] === 'name' && call[1] === 'define') ||
                       (call[0] === 'dot' && call[1][1] === 'require' && call[2] === 'def')) {

                //There should only be one argument and it should be a function,
                //or a named module with function as second arg
                if (args.length === 1 && args[0][0] === 'function') {
                    return args[0];
                } else if (args.length === 2 && args[0][0] === 'string' &&
                           args[1][0] === 'function') {
                    return args[1];
                }
            }
        }

        //Check child nodes
        for (i = 0; i < node.length; i++) {
            n = node[i];
            if ((callback = this.findAnonDefineFactory(n))) {
                return callback;
            }
        }
    }

    return null;
};

/**
 * Finds any config that is passed to requirejs.
 * @param {String} fileName
 * @param {String} fileContents
 *
 * @returns {Object} a config object. Will be null if no config.
 * Can throw an error if the config in the file cannot be evaluated in
 * a build context to valid JavaScript.
 */
parse.findConfig = function (fileName, fileContents) {
    /*jslint evil: true */
    //This is a litle bit inefficient, it ends up with two uglifyjs parser
    //calls. Can revisit later, but trying to build out larger functional
    //pieces first.
    var foundConfig = null,
        astRoot = parser.parse(fileContents);

    parse.recurse(astRoot, function (configNode) {
        var jsConfig;

        if (!foundConfig && configNode) {
            jsConfig = parse.nodeToString(configNode);
            foundConfig = eval('(' + jsConfig + ')');
            return foundConfig;
        }
        return undefined;
    }, null, parse.parseConfigNode);

    return foundConfig;
};

/**
 * Finds all dependencies specified in dependency arrays and inside
 * simplified commonjs wrappers.
 * @param {String} fileName
 * @param {String} fileContents
 *
 * @returns {Array} an array of dependency strings. The dependencies
 * have not been normalized, they may be relative IDs.
 */
parse.findDependencies = function (fileName, fileContents, options) {
    //This is a litle bit inefficient, it ends up with two uglifyjs parser
    //calls. Can revisit later, but trying to build out larger functional
    //pieces first.
    var dependencies = [],
        astRoot = parser.parse(fileContents);

    parse.recurse(astRoot, function (callName, config, name, deps) {
        //Normalize the input args.
        if (name && isArrayLiteral(name)) {
            deps = name;
            name = null;
        }

        if ((deps = getValidDeps(deps))) {
            dependencies = dependencies.concat(deps);
        }
    }, options);

    return dependencies;
};

/**
 * Finds only CJS dependencies, ones that are the form require('stringLiteral')
 */
parse.findCjsDependencies = function (fileName, fileContents, options) {
    //This is a litle bit inefficient, it ends up with two uglifyjs parser
    //calls. Can revisit later, but trying to build out larger functional
    //pieces first.
    var dependencies = [],
        astRoot = parser.parse(fileContents);

    parse.recurse(astRoot, function (dep) {
        dependencies.push(dep);
    }, options, function (node, onMatch) {

        var call, args;

        if (!isArray(node)) {
            return false;
        }

        if (node[0] === 'call') {
            call = node[1];
            args = node[2];

            if (call) {
                //A require('') use.
                if (call[0] === 'name' && call[1] === 'require' &&
                    args[0][0] === 'string') {
                    return onMatch(args[0][1]);
                }
            }
        }

        return false;

    });

    return dependencies;
};

/**
 * Determines if define(), require({}|[]) or requirejs was called in the
 * file. Also finds out if define() is declared and if define.amd is called.
 */
parse.usesAmdOrRequireJs = function (fileName, fileContents, options) {
    var astRoot = parser.parse(fileContents),
        uses;

    parse.recurse(astRoot, function (prop) {
        if (!uses) {
            uses = {};
        }
        uses[prop] = true;
    }, options, parse.findAmdOrRequireJsNode);

    return uses;
};

/**
 * Determines if require(''), exports.x =, module.exports =,
 * __dirname, __filename are used. So, not strictly traditional CommonJS,
 * also checks for Node variants.
 */
parse.usesCommonJs = function (fileName, fileContents, options) {
    var uses = null,
        assignsExports = false,
        astRoot = parser.parse(fileContents);

    parse.recurse(astRoot, function (prop) {
        if (prop === 'varExports') {
            assignsExports = true;
        } else if (prop !== 'exports' || !assignsExports) {
            if (!uses) {
                uses = {};
            }
            uses[prop] = true;
        }
    }, options, function (node, onMatch) {

        var call, args;

        if (!isArray(node)) {
            return false;
        }

        if (node[0] === 'name' && (node[1] === '__dirname' || node[1] === '__filename')) {
            return onMatch(node[1].substring(2));
        } else if (node[0] === 'var' && node[1] && node[1][0] && node[1][0][0] === 'exports') {
            //Hmm, a variable assignment for exports, so does not use cjs exports.
            return onMatch('varExports');
        } else if (node[0] === 'assign' && node[2] && node[2][0] === 'dot') {
            args = node[2][1];

            if (args) {
                //An exports or module.exports assignment.
                if (args[0] === 'name' && args[1] === 'module' &&
                    node[2][2] === 'exports') {
                    return onMatch('moduleExports');
                } else if (args[0] === 'name' && args[1] === 'exports') {
                    return onMatch('exports');
                }
            }
        } else if (node[0] === 'call') {
            call = node[1];
            args = node[2];

            if (call) {
                //A require('') use.
                if (call[0] === 'name' && call[1] === 'require' &&
                    args[0][0] === 'string') {
                    return onMatch('require');
                }
            }
        }

        return false;

    });

    return uses;
};


parse.findRequireDepNames = function (node, deps) {
    var moduleName, i, n, call, args;

    if (isArray(node)) {
        if (node[0] === 'call') {
            call = node[1];
            args = node[2];

            if (call && call[0] === 'name' && call[1] === 'require') {
                moduleName = args[0];
                if (moduleName[0] === 'string') {
                    deps.push(moduleName[1]);
                }
            }


        }

        //Check child nodes
        for (i = 0; i < node.length; i++) {
            n = node[i];
            this.findRequireDepNames(n, deps);
        }
    }
};

/**
 * Determines if a given node contains a require() definition.
 * @param {Array} node
 * @returns {Boolean}
 */
parse.nodeHasRequire = function (node) {
    if (this.isDefineNode(node)) {
        return true;
    }

    if (isArray(node)) {
        for (var i = 0, n; i < node.length; i++) {
            n = node[i];
            if (this.nodeHasRequire(n)) {
                return true;
            }
        }
    }

    return false;
};

/**
 * Is the given node the actual definition of define(). Actually uses
 * the definition of define.amd to find require.
 * @param {Array} node
 * @returns {Boolean}
 */
parse.isDefineNode = function (node) {
    //Actually look for the define.amd = assignment, since
    //that is more indicative of RequireJS vs a plain require definition.
    var assign;
    if (!node) {
        return null;
    }

    if (node[0] === 'assign' && node[1] === true) {
        assign = node[2];
        if (assign[0] === 'dot' && assign[1][0] === 'name' &&
            assign[1][1] === 'define' && assign[2] === 'amd') {
            return true;
        }
    }
    return false;
};

/**
 * Determines if a specific node is a valid require or define/require.def call.
 * @param {Array} node
 * @param {Function} onMatch a function to call when a match is found.
 * It is passed the match name, and the config, name, deps possible args.
 * The config, name and deps args are not normalized.
 *
 * @returns {String} a JS source string with the valid require/define call.
 * Otherwise null.
 */
parse.parseNode = function (node, onMatch) {
    var call, name, config, deps, args, cjsDeps;

    if (!isArray(node)) {
        return false;
    }

    if (node[0] === 'call') {
        call = node[1];
        args = node[2];

        if (call) {
            if (call[0] === 'name' &&
               (call[1] === 'require' || call[1] === 'requirejs')) {

                //It is a plain require() call.
                config = args[0];
                deps = args[1];
                if (isArrayLiteral(config)) {
                    deps = config;
                    config = null;
                }

                if (!(deps = validateDeps(deps))) {
                    return null;
                }

                return onMatch("require", null, null, deps);

            } else if (call[0] === 'name' && call[1] === 'define') {

                //A define call
                name = args[0];
                deps = args[1];
                //Only allow define calls that match what is expected
                //in an AMD call:
                //* first arg should be string, array, function or object
                //* second arg optional, or array, function or object.
                //This helps weed out calls to a non-AMD define, but it is
                //not completely robust. Someone could create a define
                //function that still matches this shape, but this is the
                //best that is possible, and at least allows UglifyJS,
                //which does create its own internal define in one file,
                //to be inlined.
                if (((name[0] === 'string' || isArrayLiteral(name) ||
                      name[0] === 'function' || isObjectLiteral(name))) &&
                    (!deps || isArrayLiteral(deps) ||
                     deps[0] === 'function' || isObjectLiteral(deps) ||
                     // allow define(['dep'], factory) pattern
                     (isArrayLiteral(name) && deps[0] === 'name' && args.length === 2))) {

                    //If first arg is a function, could be a commonjs wrapper,
                    //look inside for commonjs dependencies.
                    //Also, if deps is a function look for commonjs deps.
                    if (name && name[0] === 'function') {
                        cjsDeps = parse.getAnonDepsFromNode(name);
                        if (cjsDeps.length) {
                            name = toAstArray(cjsDeps);
                        }
                    } else if (deps && deps[0] === 'function') {
                        cjsDeps = parse.getAnonDepsFromNode(deps);
                        if (cjsDeps.length) {
                            deps = toAstArray(cjsDeps);
                        }
                    }

                    return onMatch("define", null, name, deps);
                }
            }
        }
    }

    return false;
};

/**
 * Looks for define(), require({} || []), requirejs({} || []) calls.
 */
parse.findAmdOrRequireJsNode = function (node, onMatch) {
    var call, args, configNode, type;

    if (!isArray(node)) {
        return false;
    }

    if (node[0] === 'defun' && node[1] === 'define') {
        type = 'declaresDefine';
    } else if (node[0] === 'assign' && node[2] && node[2][2] === 'amd' &&
        node[2][1] && node[2][1][0] === 'name' &&
        node[2][1][1] === 'define') {
        type = 'defineAmd';
    } else if (node[0] === 'call') {
        call = node[1];
        args = node[2];

        if (call) {
            if ((call[0] === 'dot' &&
               (call[1] && call[1][0] === 'name' &&
                (call[1][1] === 'require' || call[1][1] === 'requirejs')) &&
               call[2] === 'config')) {
                //A require.config() or requirejs.config() call.
                type = call[1][1] + 'Config';
            } else if (call[0] === 'name' &&
               (call[1] === 'require' || call[1] === 'requirejs')) {
                //A require() or requirejs() config call.
                //Only want ones that start with an object or an array.
                configNode = args[0];
                if (configNode[0] === 'object' || configNode[0] === 'array') {
                    type = call[1];
                }
            } else if (call[0] === 'name' && call[1] === 'define') {
                //A define call.
                type = 'define';
            }
        }
    }

    if (type) {
        return onMatch(type);
    }

    return false;
};

/**
 * Determines if a specific node is a valid require/requirejs config
 * call. That includes calls to require/requirejs.config().
 * @param {Array} node
 * @param {Function} onMatch a function to call when a match is found.
 * It is passed the match name, and the config, name, deps possible args.
 * The config, name and deps args are not normalized.
 *
 * @returns {String} a JS source string with the valid require/define call.
 * Otherwise null.
 */
parse.parseConfigNode = function (node, onMatch) {
    var call, configNode, args;

    if (!isArray(node)) {
        return false;
    }

    if (node[0] === 'call') {
        call = node[1];
        args = node[2];

        if (call) {
            //A require.config() or requirejs.config() call.
            if ((call[0] === 'dot' &&
               (call[1] && call[1][0] === 'name' &&
                (call[1][1] === 'require' || call[1][1] === 'requirejs')) &&
               call[2] === 'config') ||
               //A require() or requirejs() config call.

               (call[0] === 'name' &&
               (call[1] === 'require' || call[1] === 'requirejs'))
            ) {
                //It is a plain require() call.
                configNode = args[0];

                if (configNode[0] !== 'object') {
                    return null;
                }

                return onMatch(configNode);

            }
        }
    }

    return false;
};

/**
 * Converts an AST node into a JS source string. Does not maintain formatting
 * or even comments from original source, just returns valid JS source.
 * @param {Array} node
 * @returns {String} a JS source string.
 */
parse.nodeToString = function (node) {
    return processor.gen_code(node, true);
};

module.exports = parse;