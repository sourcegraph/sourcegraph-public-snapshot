var lang = require('./lang.js'),
    esprima = require('esprima');

/**
 * @license Copyright (c) 2010-2014, The Dojo Foundation All Rights Reserved.
 * Available via the MIT or new BSD license.
 * see: http://github.com/jrburke/requirejs for details
 */

function arrayToString(ary) {
    var output = '[';
    if (ary) {
        ary.forEach(function (item, i) {
            output += (i > 0 ? ',' : '') + '"' + lang.jsEscape(item) + '"';
        });
    }
    output += ']';

    return output;
}

//This string is saved off because JSLint complains
//about obj.arguments use, as 'reserved word'
var argPropName = 'arguments';

//From an esprima example for traversing its ast.
function traverse(object, visitor) {
    var key, child;

    if (!object) {
        return;
    }

    if (visitor.call(null, object) === false) {
        return false;
    }
    for (key in object) {
        if (object.hasOwnProperty(key)) {
            child = object[key];
            if (typeof child === 'object' && child !== null) {
                if (traverse(child, visitor) === false) {
                    return false;
                }
            }
        }
    }
}

//Like traverse, but visitor returning false just
//stops that subtree analysis, not the rest of tree
//visiting.
function traverseBroad(object, visitor) {
    var key, child;

    if (!object) {
        return;
    }

    if (visitor.call(null, object) === false) {
        return false;
    }
    for (key in object) {
        if (object.hasOwnProperty(key)) {
            child = object[key];
            if (typeof child === 'object' && child !== null) {
                traverse(child, visitor);
            }
        }
    }
}

/**
 * Pulls out dependencies from an array literal with just string members.
 * If string literals, will just return those string values in an array,
 * skipping other items in the array.
 *
 * @param {Node} node an AST node.
 *
 * @returns {Array} an array of strings.
 * If null is returned, then it means the input node was not a valid
 * dependency.
 */
function getValidDeps(node) {
    if (!node || node.type !== 'ArrayExpression' || !node.elements) {
        return;
    }

    var deps = [];

    node.elements.some(function (elem) {
        if (elem.type === 'Literal') {
            deps.push(elem.value);
        }
    });

    return deps.length ? deps : undefined;
}

/**
 * Main parse function. Returns a string of any valid require or
 * define/require.def calls as part of one JavaScript source string.
 * @param {String} moduleName the module name that represents this file.
 * It is used to create a default define if there is not one already for the
 * file. This allows properly tracing dependencies for builds. Otherwise, if
 * the file just has a require() call, the file dependencies will not be
 * properly reflected: the file will come before its dependencies.
 * @param {String} moduleName
 * @param {String} fileName
 * @param {String} fileContents
 * @param {Object} options optional options. insertNeedsDefine: true will
 * add calls to require.needsDefine() if appropriate.
 * @returns {String} JS source string or null, if no require or
 * define/require.def calls are found.
 */
function parse(moduleName, fileName, fileContents, options) {
    options = options || {};

    //Set up source input
    var i, moduleCall, depString,
        moduleDeps = [],
        result = '',
        moduleList = [],
        needsDefine = true,
        astRoot = esprima.parse(fileContents);

    parse.recurse(astRoot, function (callName, config, name, deps) {
        if (!deps) {
            deps = [];
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
        return !!options.findNestedDependencies;
    }, options);

    if (options.insertNeedsDefine && needsDefine) {
        result += 'require.needsDefine("' + moduleName + '");';
    }

    if (moduleDeps.length || moduleList.length) {
        for (i = 0; i < moduleList.length; i++) {
            moduleCall = moduleList[i];
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

            depString = arrayToString(moduleCall.deps);
            result += 'define("' + moduleCall.name + '",' +
                      depString + ');';
        }
        if (moduleDeps.length) {
            if (result) {
                result += '\n';
            }
            depString = arrayToString(moduleDeps);
            result += 'define("' + moduleName + '",' + depString + ');';
        }
    }

    return result || null;
}

parse.traverse = traverse;
parse.traverseBroad = traverseBroad;

/**
 * Handles parsing a file recursively for require calls.
 * @param {Array} parentNode the AST node to start with.
 * @param {Function} onMatch function to call on a parse match.
 * @param {Object} [options] This is normally the build config options if
 * it is passed.
 */
parse.recurse = function (object, onMatch, options) {
    //Like traverse, but skips if branches that would not be processed
    //after has application that results in tests of true or false boolean
    //literal values.
    var key, child,
        hasHas = options && options.has;

    if (!object) {
        return;
    }

    //If has replacement has resulted in if(true){} or if(false){}, take
    //the appropriate branch and skip the other one.
    if (hasHas && object.type === 'IfStatement' && object.test.type &&
            object.test.type === 'Literal') {
        if (object.test.value) {
            //Take the if branch
            this.recurse(object.consequent, onMatch, options);
        } else {
            //Take the else branch
            this.recurse(object.alternate, onMatch, options);
        }
    } else {
        if (this.parseNode(object, onMatch) === false) {
            return;
        }
        for (key in object) {
            if (object.hasOwnProperty(key)) {
                child = object[key];
                if (typeof child === 'object' && child !== null) {
                    this.recurse(child, onMatch, options);
                }
            }
        }
    }
};

/**
 * Determines if the file defines the require/define module API.
 * Specifically, it looks for the `define.amd = ` expression.
 * @param {String} fileName
 * @param {String} fileContents
 * @returns {Boolean}
 */
parse.definesRequire = function (fileName, fileContents) {
    var found = false;

    traverse(esprima.parse(fileContents), function (node) {
        if (parse.hasDefineAmd(node)) {
            found = true;

            //Stop traversal
            return false;
        }
    });

    return found;
};

/**
 * Finds require("") calls inside a CommonJS anonymous module wrapped in a
 * define(function(require, exports, module){}) wrapper. These dependencies
 * will be added to a modified define() call that lists the dependencies
 * on the outside of the function.
 * @param {String} fileName
 * @param {String|Object} fileContents: a string of contents, or an already
 * parsed AST tree.
 * @returns {Array} an array of module names that are dependencies. Always
 * returns an array, but could be of length zero.
 */
parse.getAnonDeps = function (fileName, fileContents) {
    var astRoot = typeof fileContents === 'string' ?
                  esprima.parse(fileContents) : fileContents,
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

        //If no deps, still add the standard CommonJS require, exports,
        //module, in that order, to the deps, but only if specified as
        //function args. In particular, if exports is used, it is favored
        //over the return value of the function, so only add it if asked.
        funcArgLength = node.params && node.params.length;
        if (funcArgLength) {
            deps = (funcArgLength > 1 ? ["require", "exports", "module"] :
                    ["require"]).concat(deps);
        }
    }
    return deps;
};

parse.isDefineNodeWithArgs = function (node) {
    return node && node.type === 'CallExpression' &&
           node.callee && node.callee.type === 'Identifier' &&
           node.callee.name === 'define' && node[argPropName];
};

/**
 * Finds the function in define(function (require, exports, module){});
 * @param {Array} node
 * @returns {Boolean}
 */
parse.findAnonDefineFactory = function (node) {
    var match;

    traverse(node, function (node) {
        var arg0, arg1;

        if (parse.isDefineNodeWithArgs(node)) {

            //Just the factory function passed to define
            arg0 = node[argPropName][0];
            if (arg0 && arg0.type === 'FunctionExpression') {
                match = arg0;
                return false;
            }

            //A string literal module ID followed by the factory function.
            arg1 = node[argPropName][1];
            if (arg0.type === 'Literal' &&
                    arg1 && arg1.type === 'FunctionExpression') {
                match = arg1;
                return false;
            }
        }
    });

    return match;
};

/**
 * Finds any config that is passed to requirejs. That includes calls to
 * require/requirejs.config(), as well as require({}, ...) and
 * requirejs({}, ...)
 * @param {String} fileContents
 *
 * @returns {Object} a config details object with the following properties:
 * - config: {Object} the config object found. Can be undefined if no
 * config found.
 * - range: {Array} the start index and end index in the contents where
 * the config was found. Can be undefined if no config found.
 * Can throw an error if the config in the file cannot be evaluated in
 * a build context to valid JavaScript.
 */
parse.findConfig = function (fileContents) {
    /*jslint evil: true */
    var jsConfig, foundConfig, stringData, foundRange, quote, quoteMatch,
        quoteRegExp = /(:\s|\[\s*)(['"])/,
        astRoot = esprima.parse(fileContents, {
            loc: true
        });

    traverse(astRoot, function (node) {
        var arg,
            requireType = parse.hasRequire(node);

        if (requireType && (requireType === 'require' ||
                requireType === 'requirejs' ||
                requireType === 'requireConfig' ||
                requireType === 'requirejsConfig')) {

            arg = node[argPropName] && node[argPropName][0];

            if (arg && arg.type === 'ObjectExpression') {
                stringData = parse.nodeToString(fileContents, arg);
                jsConfig = stringData.value;
                foundRange = stringData.range;
                return false;
            }
        } else {
            arg = parse.getRequireObjectLiteral(node);
            if (arg) {
                stringData = parse.nodeToString(fileContents, arg);
                jsConfig = stringData.value;
                foundRange = stringData.range;
                return false;
            }
        }
    });

    if (jsConfig) {
        // Eval the config
        quoteMatch = quoteRegExp.exec(jsConfig);
        quote = (quoteMatch && quoteMatch[2]) || '"';
        foundConfig = eval('(' + jsConfig + ')');
    }

    return {
        config: foundConfig,
        range: foundRange,
        quote: quote
    };
};

/** Returns the node for the object literal assigned to require/requirejs,
 * for holding a declarative config.
 */
parse.getRequireObjectLiteral = function (node) {
    if (node.id && node.id.type === 'Identifier' &&
            (node.id.name === 'require' || node.id.name === 'requirejs') &&
            node.init && node.init.type === 'ObjectExpression') {
        return node.init;
    }
};

/**
 * Renames require/requirejs/define calls to be ns + '.' + require/requirejs/define
 * Does *not* do .config calls though. See pragma.namespace for the complete
 * set of namespace transforms. This function is used because require calls
 * inside a define() call should not be renamed, so a simple regexp is not
 * good enough.
 * @param  {String} fileContents the contents to transform.
 * @param  {String} ns the namespace, *not* including trailing dot.
 * @return {String} the fileContents with the namespace applied
 */
parse.renameNamespace = function (fileContents, ns) {
    var lines,
        locs = [],
        astRoot = esprima.parse(fileContents, {
            loc: true
        });

    parse.recurse(astRoot, function (callName, config, name, deps, node) {
        locs.push(node.loc);
        //Do not recurse into define functions, they should be using
        //local defines.
        return callName !== 'define';
    }, {});

    if (locs.length) {
        lines = fileContents.split('\n');

        //Go backwards through the found locs, adding in the namespace name
        //in front.
        locs.reverse();
        locs.forEach(function (loc) {
            var startIndex = loc.start.column,
            //start.line is 1-based, not 0 based.
            lineIndex = loc.start.line - 1,
            line = lines[lineIndex];

            lines[lineIndex] = line.substring(0, startIndex) +
                               ns + '.' +
                               line.substring(startIndex,
                                                  line.length);
        });

        fileContents = lines.join('\n');
    }

    return fileContents;
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
    var dependencies = [],
        astRoot = esprima.parse(fileContents);

    parse.recurse(astRoot, function (callName, config, name, deps) {
        if (deps) {
            dependencies = dependencies.concat(deps);
        }
    }, options);

    return dependencies;
};

/**
 * Finds only CJS dependencies, ones that are the form
 * require('stringLiteral')
 */
parse.findCjsDependencies = function (fileName, fileContents) {
    var dependencies = [];

    traverse(esprima.parse(fileContents), function (node) {
        var arg;

        if (node && node.type === 'CallExpression' && node.callee &&
                node.callee.type === 'Identifier' &&
                node.callee.name === 'require' && node[argPropName] &&
                node[argPropName].length === 1) {
            arg = node[argPropName][0];
            if (arg.type === 'Literal') {
                dependencies.push(arg.value);
            }
        }
    });

    return dependencies;
};

//function define() {}
parse.hasDefDefine = function (node) {
    return node.type === 'FunctionDeclaration' && node.id &&
                node.id.type === 'Identifier' && node.id.name === 'define';
};

//define.amd = ...
parse.hasDefineAmd = function (node) {
    return node && node.type === 'AssignmentExpression' &&
        node.left && node.left.type === 'MemberExpression' &&
        node.left.object && node.left.object.name === 'define' &&
        node.left.property && node.left.property.name === 'amd';
};

//define.amd reference, as in: if (define.amd)
parse.refsDefineAmd = function (node) {
    return node && node.type === 'MemberExpression' &&
    node.object && node.object.name === 'define' &&
    node.object.type === 'Identifier' &&
    node.property && node.property.name === 'amd' &&
    node.property.type === 'Identifier';
};

//require(), requirejs(), require.config() and requirejs.config()
parse.hasRequire = function (node) {
    var callName,
        c = node && node.callee;

    if (node && node.type === 'CallExpression' && c) {
        if (c.type === 'Identifier' &&
                (c.name === 'require' ||
                c.name === 'requirejs')) {
            //A require/requirejs({}, ...) call
            callName = c.name;
        } else if (c.type === 'MemberExpression' &&
                c.object &&
                c.object.type === 'Identifier' &&
                (c.object.name === 'require' ||
                    c.object.name === 'requirejs') &&
                c.property && c.property.name === 'config') {
            // require/requirejs.config({}) call
            callName = c.object.name + 'Config';
        }
    }

    return callName;
};

//define()
parse.hasDefine = function (node) {
    return node && node.type === 'CallExpression' && node.callee &&
        node.callee.type === 'Identifier' &&
        node.callee.name === 'define';
};

/**
 * If there is a named define in the file, returns the name. Does not
 * scan for mulitple names, just the first one.
 */
parse.getNamedDefine = function (fileContents) {
    var name;
    traverse(esprima.parse(fileContents), function (node) {
        if (node && node.type === 'CallExpression' && node.callee &&
        node.callee.type === 'Identifier' &&
        node.callee.name === 'define' &&
        node[argPropName] && node[argPropName][0] &&
        node[argPropName][0].type === 'Literal') {
            name = node[argPropName][0].value;
            return false;
        }
    });

    return name;
};

/**
 * Determines if define(), require({}|[]) or requirejs was called in the
 * file. Also finds out if define() is declared and if define.amd is called.
 */
parse.usesAmdOrRequireJs = function (fileName, fileContents) {
    var uses;

    traverse(esprima.parse(fileContents), function (node) {
        var type, callName, arg;

        if (parse.hasDefDefine(node)) {
            //function define() {}
            type = 'declaresDefine';
        } else if (parse.hasDefineAmd(node)) {
            type = 'defineAmd';
        } else {
            callName = parse.hasRequire(node);
            if (callName) {
                arg = node[argPropName] && node[argPropName][0];
                if (arg && (arg.type === 'ObjectExpression' ||
                        arg.type === 'ArrayExpression')) {
                    type = callName;
                }
            } else if (parse.hasDefine(node)) {
                type = 'define';
            }
        }

        if (type) {
            if (!uses) {
                uses = {};
            }
            uses[type] = true;
        }
    });

    return uses;
};

/**
 * Determines if require(''), exports.x =, module.exports =,
 * __dirname, __filename are used. So, not strictly traditional CommonJS,
 * also checks for Node variants.
 */
parse.usesCommonJs = function (fileName, fileContents) {
    var uses = null,
        assignsExports = false;


    traverse(esprima.parse(fileContents), function (node) {
        var type,
            exp = node.expression || node.init;

        if (node.type === 'Identifier' &&
                (node.name === '__dirname' || node.name === '__filename')) {
            type = node.name.substring(2);
        } else if (node.type === 'VariableDeclarator' && node.id &&
                node.id.type === 'Identifier' &&
                    node.id.name === 'exports') {
            //Hmm, a variable assignment for exports, so does not use cjs
            //exports.
            type = 'varExports';
        } else if (exp && exp.type === 'AssignmentExpression' && exp.left &&
                exp.left.type === 'MemberExpression' && exp.left.object) {
            if (exp.left.object.name === 'module' && exp.left.property &&
                    exp.left.property.name === 'exports') {
                type = 'moduleExports';
            } else if (exp.left.object.name === 'exports' &&
                    exp.left.property) {
                type = 'exports';
            }

        } else if (node && node.type === 'CallExpression' && node.callee &&
                node.callee.type === 'Identifier' &&
                node.callee.name === 'require' && node[argPropName] &&
                node[argPropName].length === 1 &&
                node[argPropName][0].type === 'Literal') {
            type = 'require';
        }

        if (type) {
            if (type === 'varExports') {
                assignsExports = true;
            } else if (type !== 'exports' || !assignsExports) {
                if (!uses) {
                    uses = {};
                }
                uses[type] = true;
            }
        }
    });

    return uses;
};


parse.findRequireDepNames = function (node, deps) {
    traverse(node, function (node) {
        var arg;

        if (node && node.type === 'CallExpression' && node.callee &&
                node.callee.type === 'Identifier' &&
                node.callee.name === 'require' &&
                node[argPropName] && node[argPropName].length === 1) {

            arg = node[argPropName][0];
            if (arg.type === 'Literal') {
                deps.push(arg.value);
            }
        }
    });
};

/**
 * Determines if a specific node is a valid require or define/require.def
 * call.
 * @param {Array} node
 * @param {Function} onMatch a function to call when a match is found.
 * It is passed the match name, and the config, name, deps possible args.
 * The config, name and deps args are not normalized.
 *
 * @returns {String} a JS source string with the valid require/define call.
 * Otherwise null.
 */
parse.parseNode = function (node, onMatch) {
    var name, deps, cjsDeps, arg, factory, exp, refsDefine, bodyNode,
        args = node && node[argPropName],
        callName = parse.hasRequire(node);

    if (callName === 'require' || callName === 'requirejs') {
        //A plain require/requirejs call
        arg = node[argPropName] && node[argPropName][0];
        if (arg.type !== 'ArrayExpression') {
            if (arg.type === 'ObjectExpression') {
                //A config call, try the second arg.
                arg = node[argPropName][1];
            }
        }

        deps = getValidDeps(arg);
        if (!deps) {
            return;
        }

        return onMatch("require", null, null, deps, node);
    } else if (parse.hasDefine(node) && args && args.length) {
        name = args[0];
        deps = args[1];
        factory = args[2];

        if (name.type === 'ArrayExpression') {
            //No name, adjust args
            factory = deps;
            deps = name;
            name = null;
        } else if (name.type === 'FunctionExpression') {
            //Just the factory, no name or deps
            factory = name;
            name = deps = null;
        } else if (name.type !== 'Literal') {
             //An object literal, just null out
            name = deps = factory = null;
        }

        if (name && name.type === 'Literal' && deps) {
            if (deps.type === 'FunctionExpression') {
                //deps is the factory
                factory = deps;
                deps = null;
            } else if (deps.type === 'ObjectExpression') {
                //deps is object literal, null out
                deps = factory = null;
            } else if (deps.type === 'Identifier' && args.length === 2) {
                // define('id', factory)
                deps = factory = null;
            }
        }

        if (deps && deps.type === 'ArrayExpression') {
            deps = getValidDeps(deps);
        } else if (factory && factory.type === 'FunctionExpression') {
            //If no deps and a factory function, could be a commonjs sugar
            //wrapper, scan the function for dependencies.
            cjsDeps = parse.getAnonDepsFromNode(factory);
            if (cjsDeps.length) {
                deps = cjsDeps;
            }
        } else if (deps || factory) {
            //Does not match the shape of an AMD call.
            return;
        }

        //Just save off the name as a string instead of an AST object.
        if (name && name.type === 'Literal') {
            name = name.value;
        }

        return onMatch("define", null, name, deps, node);
    } else if (node.type === 'CallExpression' && node.callee &&
               node.callee.type === 'FunctionExpression' &&
               node.callee.body && node.callee.body.body &&
               node.callee.body.body.length === 1 &&
               node.callee.body.body[0].type === 'IfStatement') {
        bodyNode = node.callee.body.body[0];
        //Look for a define(Identifier) case, but only if inside an
        //if that has a define.amd test
        if (bodyNode.consequent && bodyNode.consequent.body) {
            exp = bodyNode.consequent.body[0];
            if (exp.type === 'ExpressionStatement' && exp.expression &&
                parse.hasDefine(exp.expression) &&
                exp.expression.arguments &&
                exp.expression.arguments.length === 1 &&
                exp.expression.arguments[0].type === 'Identifier') {

                //Calls define(Identifier) as first statement in body.
                //Confirm the if test references define.amd
                traverse(bodyNode.test, function (node) {
                    if (parse.refsDefineAmd(node)) {
                        refsDefine = true;
                        return false;
                    }
                });

                if (refsDefine) {
                    return onMatch("define", null, null, null, exp.expression);
                }
            }
        }
    }
};

/**
 * Converts an AST node into a JS source string by extracting
 * the node's location from the given contents string. Assumes
 * esprima.parse() with loc was done.
 * @param {String} contents
 * @param {Object} node
 * @returns {String} a JS source string.
 */
parse.nodeToString = function (contents, node) {
    var extracted,
        loc = node.loc,
        lines = contents.split('\n'),
        firstLine = loc.start.line > 1 ?
                    lines.slice(0, loc.start.line - 1).join('\n') + '\n' :
                    '',
        preamble = firstLine +
                   lines[loc.start.line - 1].substring(0, loc.start.column);

    if (loc.start.line === loc.end.line) {
        extracted = lines[loc.start.line - 1].substring(loc.start.column,
                                                        loc.end.column);
    } else {
        extracted =  lines[loc.start.line - 1].substring(loc.start.column) +
                 '\n' +
                 lines.slice(loc.start.line, loc.end.line - 1).join('\n') +
                 '\n' +
                 lines[loc.end.line - 1].substring(0, loc.end.column);
    }

    return {
        value: extracted,
        range: [
            preamble.length,
            preamble.length + extracted.length
        ]
    };
};

/**
 * Extracts license comments from JS text.
 * @param {String} fileName
 * @param {String} contents
 * @returns {String} a string of license comments.
 */
parse.getLicenseComments = function (fileName, contents) {
    var commentNode, refNode, subNode, value, i, j,
        //xpconnect's Reflect does not support comment or range, but
        //prefer continued operation vs strict parity of operation,
        //as license comments can be expressed in other ways, like
        //via wrap args, or linked via sourcemaps.
        ast = esprima.parse(contents, {
            comment: true,
            range: true
        }),
        result = '',
        existsMap = {},
        lineEnd = contents.indexOf('\r') === -1 ? '\n' : '\r\n';

    if (ast.comments) {
        for (i = 0; i < ast.comments.length; i++) {
            commentNode = ast.comments[i];

            if (commentNode.type === 'Line') {
                value = '//' + commentNode.value + lineEnd;
                refNode = commentNode;

                if (i + 1 >= ast.comments.length) {
                    value += lineEnd;
                } else {
                    //Look for immediately adjacent single line comments
                    //since it could from a multiple line comment made out
                    //of single line comments. Like this comment.
                    for (j = i + 1; j < ast.comments.length; j++) {
                        subNode = ast.comments[j];
                        if (subNode.type === 'Line' &&
                                subNode.range[0] === refNode.range[1] + 1) {
                            //Adjacent single line comment. Collect it.
                            value += '//' + subNode.value + lineEnd;
                            refNode = subNode;
                        } else {
                            //No more single line comment blocks. Break out
                            //and continue outer looping.
                            break;
                        }
                    }
                    value += lineEnd;
                    i = j - 1;
                }
            } else {
                value = '/*' + commentNode.value + '*/' + lineEnd + lineEnd;
            }

            if (!existsMap[value] && (value.indexOf('license') !== -1 ||
                    (commentNode.type === 'Block' &&
                        value.indexOf('/*!') === 0) ||
                    value.indexOf('opyright') !== -1 ||
                    value.indexOf('(c)') !== -1)) {

                result += value;
                existsMap[value] = true;
            }

        }
    }

    return result;
};

module.exports = parse;
