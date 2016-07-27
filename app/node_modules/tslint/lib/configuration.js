"use strict";
var findup = require("findup-sync");
var fs = require("fs");
var path = require("path");
var resolve = require("resolve");
var utils_1 = require("./utils");
exports.CONFIG_FILENAME = "tslint.json";
exports.DEFAULT_CONFIG = {
    "rules": {
        "class-name": true,
        "comment-format": [true, "check-space"],
        "indent": [true, "spaces"],
        "no-duplicate-variable": true,
        "no-eval": true,
        "no-internal-module": true,
        "no-trailing-whitespace": true,
        "no-unsafe-finally": true,
        "no-var-keyword": true,
        "one-line": [true, "check-open-brace", "check-whitespace"],
        "quotemark": [true, "double"],
        "semicolon": [true, "always"],
        "triple-equals": [true, "allow-null-check"],
        "typedef-whitespace": [
            true, {
                "call-signature": "nospace",
                "index-signature": "nospace",
                "parameter": "nospace",
                "property-declaration": "nospace",
                "variable-declaration": "nospace",
            },
        ],
        "variable-name": [true, "ban-keywords"],
        "whitespace": [true,
            "check-branch",
            "check-decl",
            "check-operator",
            "check-separator",
            "check-type",
        ],
    },
};
var PACKAGE_DEPRECATION_MSG = "Configuration of TSLint via package.json has been deprecated, "
    + "please start using a tslint.json file instead (http://palantir.github.io/tslint/usage/tslint-json/).";
var BUILT_IN_CONFIG = /^tslint:(.*)$/;
function findConfiguration(configFile, inputFilePath) {
    var configPath = findConfigurationPath(configFile, inputFilePath);
    return loadConfigurationFromPath(configPath);
}
exports.findConfiguration = findConfiguration;
function findConfigurationPath(suppliedConfigFilePath, inputFilePath) {
    if (suppliedConfigFilePath != null) {
        if (!fs.existsSync(suppliedConfigFilePath)) {
            throw new Error("Could not find config file at: " + path.resolve(suppliedConfigFilePath));
        }
        else {
            return path.resolve(suppliedConfigFilePath);
        }
    }
    else {
        var configFilePath = findup(exports.CONFIG_FILENAME, { cwd: inputFilePath, nocase: true });
        if (configFilePath != null && fs.existsSync(configFilePath)) {
            return path.resolve(configFilePath);
        }
        configFilePath = findup("package.json", { cwd: inputFilePath, nocase: true });
        if (configFilePath != null && require(configFilePath).tslintConfig != null) {
            return path.resolve(configFilePath);
        }
        var homeDir = getHomeDir();
        if (homeDir != null) {
            configFilePath = path.join(homeDir, exports.CONFIG_FILENAME);
            if (fs.existsSync(configFilePath)) {
                return path.resolve(configFilePath);
            }
        }
        return undefined;
    }
}
exports.findConfigurationPath = findConfigurationPath;
function loadConfigurationFromPath(configFilePath) {
    if (configFilePath == null) {
        return exports.DEFAULT_CONFIG;
    }
    else if (path.basename(configFilePath) === "package.json") {
        console.warn(PACKAGE_DEPRECATION_MSG);
        return require(configFilePath).tslintConfig;
    }
    else {
        var resolvedConfigFilePath = resolveConfigurationPath(configFilePath);
        var configFile = void 0;
        if (path.extname(resolvedConfigFilePath) === ".json") {
            var fileContent = utils_1.stripComments(fs.readFileSync(resolvedConfigFilePath)
                .toString()
                .replace(/^\uFEFF/, ""));
            configFile = JSON.parse(fileContent);
        }
        else {
            configFile = require(resolvedConfigFilePath);
            delete require.cache[resolvedConfigFilePath];
        }
        var configFileDir = path.dirname(resolvedConfigFilePath);
        configFile.rulesDirectory = getRulesDirectories(configFile.rulesDirectory, configFileDir);
        configFile.extends = utils_1.arrayify(configFile.extends);
        for (var _i = 0, _a = configFile.extends; _i < _a.length; _i++) {
            var name_1 = _a[_i];
            var baseConfigFilePath = resolveConfigurationPath(name_1, configFileDir);
            var baseConfigFile = loadConfigurationFromPath(baseConfigFilePath);
            configFile = extendConfigurationFile(configFile, baseConfigFile);
        }
        return configFile;
    }
}
exports.loadConfigurationFromPath = loadConfigurationFromPath;
function resolveConfigurationPath(filePath, relativeTo) {
    var matches = filePath.match(BUILT_IN_CONFIG);
    var isBuiltInConfig = matches != null;
    if (isBuiltInConfig) {
        var configName = matches[1];
        try {
            return require.resolve("./configs/" + configName);
        }
        catch (err) {
            throw new Error(filePath + " is not a built-in config, try \"tslint:recommended\" instead.");
        }
    }
    var basedir = relativeTo || process.cwd();
    try {
        return resolve.sync(filePath, { basedir: basedir });
    }
    catch (err) {
        try {
            return require.resolve(filePath);
        }
        catch (err) {
            throw new Error(("Invalid \"extends\" configuration value - could not require \"" + filePath + "\". ") +
                "Review the Node lookup algorithm (https://nodejs.org/api/modules.html#modules_all_together) " +
                "for the approximate method TSLint uses to find the referenced configuration file.");
        }
    }
}
function extendConfigurationFile(config, baseConfig) {
    var combinedConfig = {};
    var baseRulesDirectory = utils_1.arrayify(baseConfig.rulesDirectory);
    var configRulesDirectory = utils_1.arrayify(config.rulesDirectory);
    combinedConfig.rulesDirectory = configRulesDirectory.concat(baseRulesDirectory);
    combinedConfig.rules = {};
    for (var _i = 0, _a = Object.keys(utils_1.objectify(baseConfig.rules)); _i < _a.length; _i++) {
        var name_2 = _a[_i];
        combinedConfig.rules[name_2] = baseConfig.rules[name_2];
    }
    for (var _b = 0, _c = Object.keys(utils_1.objectify(config.rules)); _b < _c.length; _b++) {
        var name_3 = _c[_b];
        combinedConfig.rules[name_3] = config.rules[name_3];
    }
    return combinedConfig;
}
exports.extendConfigurationFile = extendConfigurationFile;
function getHomeDir() {
    var environment = global.process.env;
    var paths = [
        environment.USERPROFILE,
        environment.HOME,
        environment.HOMEPATH,
        environment.HOMEDRIVE + environment.HOMEPATH,
    ];
    for (var _i = 0, paths_1 = paths; _i < paths_1.length; _i++) {
        var homePath = paths_1[_i];
        if (homePath != null && fs.existsSync(homePath)) {
            return homePath;
        }
    }
}
function getRelativePath(directory, relativeTo) {
    if (directory != null) {
        var basePath = relativeTo || process.cwd();
        return path.resolve(basePath, directory);
    }
}
exports.getRelativePath = getRelativePath;
function getRulesDirectories(directories, relativeTo) {
    var rulesDirectories = utils_1.arrayify(directories).map(function (dir) { return getRelativePath(dir, relativeTo); });
    for (var _i = 0, rulesDirectories_1 = rulesDirectories; _i < rulesDirectories_1.length; _i++) {
        var directory = rulesDirectories_1[_i];
        if (!fs.existsSync(directory)) {
            throw new Error("Could not find custom rule directory: " + directory);
        }
    }
    return rulesDirectories;
}
exports.getRulesDirectories = getRulesDirectories;
