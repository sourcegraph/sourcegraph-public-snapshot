"use strict";
var objectAssign = require('object-assign');
var path = require('path');
var utils = require('./utils');
function getConfigFile(compiler, loader, loaderOptions, compilerCompatible, log, compilerDetailsLogMessage) {
    var configFilePath = findConfigFile(compiler, path.dirname(loader.resourcePath), loaderOptions.configFileName);
    var configFileError;
    var configFile;
    if (configFilePath) {
        if (compilerCompatible) {
            log.logInfo((compilerDetailsLogMessage + " and " + configFilePath).green);
        }
        else {
            log.logInfo(("ts-loader: Using config file at " + configFilePath).green);
        }
        // HACK: relies on the fact that passing an extra argument won't break
        // the old API that has a single parameter
        configFile = compiler.readConfigFile(configFilePath, compiler.sys.readFile);
        if (configFile.error) {
            configFileError = utils.formatErrors([configFile.error], loaderOptions, compiler, { file: configFilePath })[0];
        }
    }
    else {
        if (compilerCompatible) {
            log.logInfo(compilerDetailsLogMessage.green);
        }
        configFile = {
            config: {
                compilerOptions: {},
                files: []
            }
        };
    }
    if (!configFileError) {
        configFile.config.compilerOptions = objectAssign({}, configFile.config.compilerOptions, loaderOptions.compilerOptions);
        // do any necessary config massaging
        if (loaderOptions.transpileOnly) {
            configFile.config.compilerOptions.isolatedModules = true;
        }
    }
    return {
        configFilePath: configFilePath,
        configFile: configFile,
        configFileError: configFileError
    };
}
exports.getConfigFile = getConfigFile;
/**
 * The tsconfig.json is found using the same method as `tsc`, starting in the current directory
 * and continuing up the parent directory chain.
 */
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
function getConfigParseResult(compiler, configFile, configFilePath) {
    var configParseResult;
    if (typeof compiler.parseJsonConfigFileContent === 'function') {
        // parseConfigFile was renamed between 1.6.2 and 1.7
        configParseResult = compiler.parseJsonConfigFileContent(configFile.config, compiler.sys, path.dirname(configFilePath || ''));
    }
    else {
        configParseResult = compiler.parseConfigFile(configFile.config, compiler.sys, path.dirname(configFilePath || ''));
    }
    return configParseResult;
}
exports.getConfigParseResult = getConfigParseResult;
