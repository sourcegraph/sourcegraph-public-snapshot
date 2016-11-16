"use strict";
var objectAssign = require('object-assign');
var semver = require('semver');
function getCompiler(loaderOptions, log) {
    var compiler;
    var errorMessage;
    var compilerDetailsLogMessage;
    var compilerCompatible = false;
    try {
        compiler = require(loaderOptions.compiler);
    }
    catch (e) {
        errorMessage = loaderOptions.compiler === 'typescript'
            ? 'Could not load TypeScript. Try installing with `npm install typescript`. If TypeScript is installed globally, try using `npm link typescript`.'
            : "Could not load TypeScript compiler with NPM package name `" + loaderOptions.compiler + "`. Are you sure it is correctly installed?";
    }
    if (!errorMessage) {
        compilerDetailsLogMessage = "ts-loader: Using " + loaderOptions.compiler + "@" + compiler.version;
        compilerCompatible = false;
        if (loaderOptions.compiler === 'typescript') {
            if (compiler.version && semver.gte(compiler.version, '1.6.2-0')) {
                // don't log yet in this case, if a tsconfig.json exists we want to combine the message
                compilerCompatible = true;
            }
            else {
                log.logError((compilerDetailsLogMessage + ". This version is incompatible with ts-loader. Please upgrade to the latest version of TypeScript.").red);
            }
        }
        else {
            log.logWarning((compilerDetailsLogMessage + ". This version may or may not be compatible with ts-loader.").yellow);
        }
    }
    return { compiler: compiler, compilerCompatible: compilerCompatible, compilerDetailsLogMessage: compilerDetailsLogMessage, errorMessage: errorMessage };
}
exports.getCompiler = getCompiler;
function getCompilerOptions(compilerCompatible, compiler, configParseResult) {
    var compilerOptions = objectAssign({}, configParseResult.options, {
        skipDefaultLibCheck: true,
        suppressOutputPathCheck: true
    });
    // if `module` is not specified and not using ES6 target, default to CJS module output
    if ((!compilerOptions.module) && compilerOptions.target !== 2 /* ES6 */) {
        compilerOptions.module = 1; /* CommonJS */
    }
    else if (compilerCompatible && semver.lt(compiler.version, '1.7.3-0') && compilerOptions.target === 2 /* ES6 */) {
        // special handling for TS 1.6 and target: es6
        compilerOptions.module = 0 /* None */;
    }
    return compilerOptions;
}
exports.getCompilerOptions = getCompilerOptions;
