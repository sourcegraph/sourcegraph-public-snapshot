"use strict";
var path = require('path');
var fs = require('fs');
var os = require('os');
var loaderUtils = require('loader-utils');
var objectAssign = require('object-assign');
var arrify = require('arrify');
var makeResolver = require('./resolver');
var Console = require('console').Console;
var semver = require('semver');
require('colors');
var console = new Console(process.stderr);
var pushArray = function (arr, toPush) {
    Array.prototype.splice.apply(arr, [0, 0].concat(toPush));
};
function hasOwnProperty(obj, property) {
    return Object.prototype.hasOwnProperty.call(obj, property);
}
var instances = {};
var webpackInstances = [];
var scriptRegex = /\.tsx?$/i;
// Take TypeScript errors, parse them and format to webpack errors
// Optionally adds a file name
function formatErrors(diagnostics, instance, merge) {
    return diagnostics
        .filter(function (diagnostic) { return instance.loaderOptions.ignoreDiagnostics.indexOf(diagnostic.code) == -1; })
        .map(function (diagnostic) {
        var errorCategory = instance.compiler.DiagnosticCategory[diagnostic.category].toLowerCase();
        var errorCategoryAndCode = errorCategory + ' TS' + diagnostic.code + ': ';
        var messageText = errorCategoryAndCode + instance.compiler.flattenDiagnosticMessageText(diagnostic.messageText, os.EOL);
        if (diagnostic.file) {
            var lineChar = diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start);
            return {
                message: "" + '('.white + (lineChar.line + 1).toString().cyan + "," + (lineChar.character + 1).toString().cyan + "): " + messageText.red,
                rawMessage: messageText,
                location: { line: lineChar.line + 1, character: lineChar.character + 1 },
                loaderSource: 'ts-loader'
            };
        }
        else {
            return {
                message: "" + messageText.red,
                rawMessage: messageText,
                loaderSource: 'ts-loader'
            };
        }
    })
        .map(function (error) { return objectAssign(error, merge); });
}
// The tsconfig.json is found using the same method as `tsc`, starting in the current directory
// and continuing up the parent directory chain.
function findConfigFile(compiler, searchPath, configFileName) {
    while (true) {
        var fileName = path.join(searchPath, configFileName);
        if (compiler.sys.fileExists(fileName)) {
            return fileName;
        }
        var parentPath = path.dirname(searchPath);
        if (parentPath === searchPath) {
            break;
        }
        searchPath = parentPath;
    }
    return undefined;
}
// The loader is executed once for each file seen by webpack. However, we need to keep
// a persistent instance of TypeScript that contains all of the files in the program
// along with definition files and options. This function either creates an instance
// or returns the existing one. Multiple instances are possible by using the
// `instance` property.
function ensureTypeScriptInstance(loaderOptions, loader) {
    function log() {
        var messages = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            messages[_i - 0] = arguments[_i];
        }
        if (!loaderOptions.silent) {
            console.log.apply(console, messages);
        }
    }
    if (hasOwnProperty(instances, loaderOptions.instance)) {
        return { instance: instances[loaderOptions.instance] };
    }
    try {
        var compiler = require(loaderOptions.compiler);
    }
    catch (e) {
        var message = loaderOptions.compiler == 'typescript'
            ? 'Could not load TypeScript. Try installing with `npm install typescript`. If TypeScript is installed globally, try using `npm link typescript`.'
            : "Could not load TypeScript compiler with NPM package name `" + loaderOptions.compiler + "`. Are you sure it is correctly installed?";
        return { error: {
                message: message.red,
                rawMessage: message,
                loaderSource: 'ts-loader'
            } };
    }
    var motd = "ts-loader: Using " + loaderOptions.compiler + "@" + compiler.version, compilerCompatible = false;
    if (loaderOptions.compiler == 'typescript') {
        if (compiler.version && semver.gte(compiler.version, '1.6.2-0')) {
            // don't log yet in this case, if a tsconfig.json exists we want to combine the message
            compilerCompatible = true;
        }
        else {
            log((motd + ". This version is incompatible with ts-loader. Please upgrade to the latest version of TypeScript.").red);
        }
    }
    else {
        log((motd + ". This version may or may not be compatible with ts-loader.").yellow);
    }
    var files = {};
    var instance = instances[loaderOptions.instance] = {
        compiler: compiler,
        compilerOptions: null,
        loaderOptions: loaderOptions,
        files: files,
        languageService: null,
        version: 0,
        dependencyGraph: {}
    };
    var compilerOptions = {};
    // Load any available tsconfig.json file
    var filesToLoad = [];
    var configFilePath = findConfigFile(compiler, path.dirname(loader.resourcePath), loaderOptions.configFileName);
    var configFile;
    if (configFilePath) {
        if (compilerCompatible)
            log((motd + " and " + configFilePath).green);
        else
            log(("ts-loader: Using config file at " + configFilePath).green);
        // HACK: relies on the fact that passing an extra argument won't break
        // the old API that has a single parameter
        configFile = compiler.readConfigFile(configFilePath, compiler.sys.readFile);
        if (configFile.error) {
            var configFileError = formatErrors([configFile.error], instance, { file: configFilePath })[0];
            return { error: configFileError };
        }
    }
    else {
        if (compilerCompatible)
            log(motd.green);
        configFile = {
            config: {
                compilerOptions: {},
                files: []
            }
        };
    }
    configFile.config.compilerOptions = objectAssign({}, configFile.config.compilerOptions, loaderOptions.compilerOptions);
    // do any necessary config massaging
    if (loaderOptions.transpileOnly) {
        configFile.config.compilerOptions.isolatedModules = true;
    }
    var configParseResult;
    if (typeof compiler.parseJsonConfigFileContent === 'function') {
        // parseConfigFile was renamed between 1.6.2 and 1.7
        configParseResult = compiler.parseJsonConfigFileContent(configFile.config, compiler.sys, path.dirname(configFilePath));
    }
    else {
        configParseResult = compiler.parseConfigFile(configFile.config, compiler.sys, path.dirname(configFilePath));
    }
    if (configParseResult.errors.length) {
        pushArray(loader._module.errors, formatErrors(configParseResult.errors, instance, { file: configFilePath }));
        return { error: {
                file: configFilePath,
                message: 'error while parsing tsconfig.json'.red,
                rawMessage: 'error while parsing tsconfig.json',
                loaderSource: 'ts-loader'
            } };
    }
    instance.compilerOptions = objectAssign(compilerOptions, configParseResult.options);
    filesToLoad = configParseResult.fileNames;
    // if `module` is not specified and not using ES6 target, default to CJS module output
    if (compilerOptions.module == null && compilerOptions.target !== 2 /* ES6 */) {
        compilerOptions.module = 1; /* CommonJS */
    }
    else if (compilerCompatible && semver.lt(compiler.version, '1.7.3-0') && compilerOptions.target == 2 /* ES6 */) {
        compilerOptions.module = 0 /* None */;
    }
    if (loaderOptions.transpileOnly) {
        // quick return for transpiling
        // we do need to check for any issues with TS options though
        var program = compiler.createProgram([], compilerOptions), diagnostics = program.getOptionsDiagnostics();
        pushArray(loader._module.errors, formatErrors(diagnostics, instance, { file: configFilePath || 'tsconfig.json' }));
        return { instance: instances[loaderOptions.instance] = { compiler: compiler, compilerOptions: compilerOptions, loaderOptions: loaderOptions, files: files, dependencyGraph: {} } };
    }
    // Load initial files (core lib files, any files specified in tsconfig.json)
    var filePath;
    try {
        filesToLoad.forEach(function (fp) {
            filePath = path.normalize(fp);
            files[filePath] = {
                text: fs.readFileSync(filePath, 'utf-8'),
                version: 0
            };
        });
    }
    catch (exc) {
        var filePathError = "A file specified in tsconfig.json could not be found: " + filePath;
        return { error: {
                message: filePathError.red,
                rawMessage: filePathError,
                loaderSource: 'ts-loader'
            } };
    }
    var newLine = compilerOptions.newLine === 0 /* CarriageReturnLineFeed */ ? '\r\n' :
        compilerOptions.newLine === 1 /* LineFeed */ ? '\n' :
            os.EOL;
    // make a (sync) resolver that follows webpack's rules
    var resolver = makeResolver(loader.options);
    var readFile = function (fileName) {
        fileName = path.normalize(fileName);
        try {
            return fs.readFileSync(fileName, { encoding: 'utf8' });
        }
        catch (e) {
            return;
        }
    };
    var moduleResolutionHost = {
        fileExists: function (fileName) { return readFile(fileName) !== undefined; },
        readFile: function (fileName) { return readFile(fileName); }
    };
    // Create the TypeScript language service
    var servicesHost = {
        getProjectVersion: function () { return instance.version + ''; },
        getScriptFileNames: function () { return Object.keys(files).filter(function (filePath) { return scriptRegex.test(filePath); }); },
        getScriptVersion: function (fileName) {
            fileName = path.normalize(fileName);
            return files[fileName] && files[fileName].version.toString();
        },
        getScriptSnapshot: function (fileName) {
            // This is called any time TypeScript needs a file's text
            // We either load from memory or from disk
            fileName = path.normalize(fileName);
            var file = files[fileName];
            if (!file) {
                var text = readFile(fileName);
                if (text == null)
                    return;
                file = files[fileName] = { version: 0, text: text };
            }
            return compiler.ScriptSnapshot.fromString(file.text);
        },
        getCurrentDirectory: function () { return process.cwd(); },
        getCompilationSettings: function () { return compilerOptions; },
        getDefaultLibFileName: function (options) { return compiler.getDefaultLibFilePath(options); },
        getNewLine: function () { return newLine; },
        log: log,
        resolveModuleNames: function (moduleNames, containingFile) {
            var resolvedModules = [];
            for (var _i = 0, moduleNames_1 = moduleNames; _i < moduleNames_1.length; _i++) {
                var moduleName = moduleNames_1[_i];
                var resolvedFileName = void 0;
                var resolutionResult = void 0;
                try {
                    resolvedFileName = resolver.resolveSync(path.normalize(path.dirname(containingFile)), moduleName);
                    if (!resolvedFileName.match(/\.tsx?$/))
                        resolvedFileName = null;
                    else
                        resolutionResult = { resolvedFileName: resolvedFileName };
                }
                catch (e) {
                    resolvedFileName = null;
                }
                var tsResolution = compiler.resolveModuleName(moduleName, containingFile, compilerOptions, moduleResolutionHost);
                if (tsResolution.resolvedModule) {
                    if (resolvedFileName) {
                        if (resolvedFileName == tsResolution.resolvedModule.resolvedFileName) {
                            resolutionResult.isExternalLibraryImport = tsResolution.resolvedModule.isExternalLibraryImport;
                        }
                    }
                    else
                        resolutionResult = tsResolution.resolvedModule;
                }
                resolvedModules.push(resolutionResult);
            }
            instance.dependencyGraph[containingFile] = resolvedModules.filter(function (m) { return m != null; }).map(function (m) { return m.resolvedFileName; });
            return resolvedModules;
        }
    };
    var languageService = instance.languageService = compiler.createLanguageService(servicesHost, compiler.createDocumentRegistry());
    var getCompilerOptionDiagnostics = true;
    loader._compiler.plugin("after-compile", function (compilation, callback) {
        // Don't add errors for child compilations
        if (compilation.compiler.isChild()) {
            callback();
            return;
        }
        var stats = compilation.stats;
        // handle all other errors. The basic approach here to get accurate error
        // reporting is to start with a "blank slate" each compilation and gather
        // all errors from all files. Since webpack tracks errors in a module from
        // compilation-to-compilation, and since not every module always runs through
        // the loader, we need to detect and remove any pre-existing errors.
        function removeTSLoaderErrors(errors) {
            var index = -1, length = errors.length;
            while (++index < length) {
                if (errors[index].loaderSource == 'ts-loader') {
                    errors.splice(index--, 1);
                    length--;
                }
            }
        }
        removeTSLoaderErrors(compilation.errors);
        // handle compiler option errors after the first compile
        if (getCompilerOptionDiagnostics) {
            getCompilerOptionDiagnostics = false;
            pushArray(compilation.errors, formatErrors(languageService.getCompilerOptionsDiagnostics(), instance, { file: configFilePath || 'tsconfig.json' }));
        }
        // build map of all modules based on normalized filename
        // this is used for quick-lookup when trying to find modules
        // based on filepath
        var modules = {};
        compilation.modules.forEach(function (module) {
            if (module.resource) {
                var modulePath = path.normalize(module.resource);
                if (hasOwnProperty(modules, modulePath)) {
                    var existingModules = modules[modulePath];
                    if (existingModules.indexOf(module) == -1) {
                        existingModules.push(module);
                    }
                }
                else {
                    modules[modulePath] = [module];
                }
            }
        });
        // gather all errors from TypeScript and output them to webpack
        Object.keys(instance.files)
            .filter(function (filePath) { return !!filePath.match(/(\.d)?\.ts(x?)$/); })
            .forEach(function (filePath) {
            var errors = languageService.getSyntacticDiagnostics(filePath).concat(languageService.getSemanticDiagnostics(filePath));
            // if we have access to a webpack module, use that
            if (hasOwnProperty(modules, filePath)) {
                var associatedModules = modules[filePath];
                associatedModules.forEach(function (module) {
                    // remove any existing errors
                    removeTSLoaderErrors(module.errors);
                    // append errors
                    var formattedErrors = formatErrors(errors, instance, { module: module });
                    pushArray(module.errors, formattedErrors);
                    pushArray(compilation.errors, formattedErrors);
                });
            }
            else {
                pushArray(compilation.errors, formatErrors(errors, instance, { file: filePath }));
            }
        });
        // gather all declaration files from TypeScript and output them to webpack
        Object.keys(instance.files)
            .filter(function (filePath) { return !!filePath.match(/\.ts(x?)$/); })
            .forEach(function (filePath) {
            var output = languageService.getEmitOutput(filePath);
            var declarationFile = output.outputFiles.filter(function (filePath) { return !!filePath.name.match(/\.d.ts$/); }).pop();
            if (declarationFile) {
                compilation.assets[declarationFile.name] = {
                    source: function () { return declarationFile.text; },
                    size: function () { return declarationFile.text.length; }
                };
            }
        });
        callback();
    });
    // manually update changed files
    loader._compiler.plugin("watch-run", function (watching, cb) {
        var mtimes = watching.compiler.watchFileSystem.watcher.mtimes;
        Object.keys(mtimes)
            .filter(function (filePath) { return !!filePath.match(/\.tsx?$|\.jsx?$/); })
            .forEach(function (filePath) {
            filePath = path.normalize(filePath);
            var file = instance.files[filePath];
            if (file) {
                file.text = fs.readFileSync(filePath, { encoding: 'utf8' });
                file.version++;
                instance.version++;
            }
        });
        cb();
    });
    return { instance: instance };
}
function loader(contents) {
    this.cacheable && this.cacheable();
    var callback = this.async();
    var filePath = path.normalize(this.resourcePath);
    var queryOptions = loaderUtils.parseQuery(this.query);
    var configFileOptions = this.options.ts || {};
    var options = objectAssign({}, {
        silent: false,
        instance: 'default',
        compiler: 'typescript',
        configFileName: 'tsconfig.json',
        transpileOnly: false,
        compilerOptions: {}
    }, configFileOptions, queryOptions);
    options.ignoreDiagnostics = arrify(options.ignoreDiagnostics).map(Number);
    // differentiate the TypeScript instance based on the webpack instance
    var webpackIndex = webpackInstances.indexOf(this._compiler);
    if (webpackIndex == -1) {
        webpackIndex = webpackInstances.push(this._compiler) - 1;
    }
    options.instance = webpackIndex + '_' + options.instance;
    var _a = ensureTypeScriptInstance(options, this), instance = _a.instance, error = _a.error;
    if (error) {
        callback(error);
        return;
    }
    // Update file contents
    var file = instance.files[filePath];
    if (!file) {
        file = instance.files[filePath] = { version: 0 };
    }
    if (file.text !== contents) {
        file.version++;
        file.text = contents;
        instance.version++;
    }
    var outputText, sourceMapText, diagnostics = [];
    if (options.transpileOnly) {
        var fileName = path.basename(filePath);
        var transpileResult = instance.compiler.transpileModule(contents, {
            compilerOptions: instance.compilerOptions,
            reportDiagnostics: true,
            fileName: fileName
        });
        (outputText = transpileResult.outputText, sourceMapText = transpileResult.sourceMapText, diagnostics = transpileResult.diagnostics, transpileResult);
        pushArray(this._module.errors, formatErrors(diagnostics, instance, { module: this._module }));
    }
    else {
        var langService = instance.languageService;
        // Emit Javascript
        var output = langService.getEmitOutput(filePath);
        // Make this file dependent on *all* definition files in the program
        this.clearDependencies();
        this.addDependency(filePath);
        var allDefinitionFiles = Object.keys(instance.files).filter(function (filePath) { return /\.d\.ts$/.test(filePath); });
        allDefinitionFiles.forEach(this.addDependency.bind(this));
        // Additionally make this file dependent on all imported files
        var additionalDependencies = instance.dependencyGraph[filePath];
        if (additionalDependencies) {
            additionalDependencies.forEach(this.addDependency.bind(this));
        }
        this._module.meta.tsLoaderDefinitionFileVersions = allDefinitionFiles
            .concat(additionalDependencies)
            .map(function (filePath) { return filePath + '@' + (instance.files[filePath] || { version: '?' }).version; });
        var outputFile = output.outputFiles.filter(function (file) { return !!file.name.match(/\.js(x?)$/); }).pop();
        if (outputFile) {
            outputText = outputFile.text;
        }
        var sourceMapFile = output.outputFiles.filter(function (file) { return !!file.name.match(/\.js(x?)\.map$/); }).pop();
        if (sourceMapFile) {
            sourceMapText = sourceMapFile.text;
        }
    }
    if (outputText == null)
        throw new Error("Typescript emitted no output for " + filePath);
    if (sourceMapText) {
        var sourceMap = JSON.parse(sourceMapText);
        sourceMap.sources = [loaderUtils.getRemainingRequest(this)];
        sourceMap.file = filePath;
        sourceMap.sourcesContent = [contents];
        outputText = outputText.replace(/^\/\/# sourceMappingURL=[^\r\n]*/gm, '');
    }
    // Make sure webpack is aware that even though the emitted JavaScript may be the same as
    // a previously cached version the TypeScript may be different and therefore should be
    // treated as new
    this._module.meta.tsLoaderFileVersion = file.version;
    callback(null, outputText, sourceMap);
}
module.exports = loader;
