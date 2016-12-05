"use strict";
var path = require('path');
var fs = require('fs');
require('colors');
var afterCompile = require('./after-compile');
var config = require('./config');
var compilerSetup = require('./compilerSetup');
var utils = require('./utils');
var logger = require('./logger');
var makeServicesHost = require('./servicesHost');
var watchRun = require('./watch-run');
var instances = {};
/**
 * The loader is executed once for each file seen by webpack. However, we need to keep
 * a persistent instance of TypeScript that contains all of the files in the program
 * along with definition files and options. This function either creates an instance
 * or returns the existing one. Multiple instances are possible by using the
 * `instance` property.
 */
function ensureTypeScriptInstance(loaderOptions, loader) {
    if (utils.hasOwnProperty(instances, loaderOptions.instance)) {
        return { instance: instances[loaderOptions.instance] };
    }
    var log = logger.makeLogger(loaderOptions);
    var _a = compilerSetup.getCompiler(loaderOptions, log), compiler = _a.compiler, compilerCompatible = _a.compilerCompatible, compilerDetailsLogMessage = _a.compilerDetailsLogMessage, errorMessage = _a.errorMessage;
    if (errorMessage) {
        return { error: utils.makeError({ rawMessage: errorMessage }) };
    }
    var _b = config.getConfigFile(compiler, loader, loaderOptions, compilerCompatible, log, compilerDetailsLogMessage), configFilePath = _b.configFilePath, configFile = _b.configFile, configFileError = _b.configFileError;
    if (configFileError) {
        return { error: configFileError };
    }
    var configParseResult = config.getConfigParseResult(compiler, configFile, configFilePath);
    if (configParseResult.errors.length) {
        utils.registerWebpackErrors(loader._module.errors, utils.formatErrors(configParseResult.errors, loaderOptions, compiler, { file: configFilePath }));
        return { error: utils.makeError({ rawMessage: 'error while parsing tsconfig.json', file: configFilePath }) };
    }
    var compilerOptions = compilerSetup.getCompilerOptions(compilerCompatible, compiler, configParseResult);
    var files = {};
    if (loaderOptions.transpileOnly) {
        // quick return for transpiling
        // we do need to check for any issues with TS options though
        var program = compiler.createProgram([], compilerOptions);
        var diagnostics = program.getOptionsDiagnostics();
        utils.registerWebpackErrors(loader._module.errors, utils.formatErrors(diagnostics, loaderOptions, compiler, { file: configFilePath || 'tsconfig.json' }));
        return { instance: instances[loaderOptions.instance] = { compiler: compiler, compilerOptions: compilerOptions, loaderOptions: loaderOptions, files: files, dependencyGraph: {}, reverseDependencyGraph: {} } };
    }
    // Load initial files (core lib files, any files specified in tsconfig.json)
    var filePath;
    try {
        var filesToLoad = configParseResult.fileNames;
        filesToLoad.forEach(function (fp) {
            filePath = path.normalize(fp);
            files[filePath] = {
                text: fs.readFileSync(filePath, 'utf-8'),
                version: 0
            };
        });
    }
    catch (exc) {
        return { error: utils.makeError({
                rawMessage: "A file specified in tsconfig.json could not be found: " + filePath
            }) };
    }
    // if allowJs is set then we should accept js(x) files
    var scriptRegex = configFile.config.compilerOptions.allowJs
        ? /\.tsx?$|\.jsx?$/i
        : /\.tsx?$/i;
    var instance = instances[loaderOptions.instance] = {
        compiler: compiler,
        compilerOptions: compilerOptions,
        loaderOptions: loaderOptions,
        files: files,
        languageService: null,
        version: 0,
        dependencyGraph: {},
        reverseDependencyGraph: {},
        modifiedFiles: null
    };
    var servicesHost = makeServicesHost(scriptRegex, log, loader, instance, loaderOptions.appendTsSuffixTo);
    instance.languageService = compiler.createLanguageService(servicesHost, compiler.createDocumentRegistry());
    loader._compiler.plugin("after-compile", afterCompile(instance, configFilePath));
    loader._compiler.plugin("watch-run", watchRun(instance));
    return { instance: instance };
}
exports.ensureTypeScriptInstance = ensureTypeScriptInstance;
