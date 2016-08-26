"use strict";
var fs = require("fs");
var path = require("path");
var underscore_string_1 = require("underscore.string");
var moduleDirectory = path.dirname(module.filename);
var CORE_FORMATTERS_DIRECTORY = path.resolve(moduleDirectory, ".", "formatters");
function isFunction(variable) {
    return typeof variable === "function";
}
function isString(variable) {
    return typeof variable === "string";
}
function findFormatter(name, formattersDirectory) {
    if (isFunction(name)) {
        return name;
    }
    else if (isString(name)) {
        var camelizedName = underscore_string_1.camelize(name + "Formatter");
        var Formatter = loadFormatter(CORE_FORMATTERS_DIRECTORY, camelizedName);
        if (Formatter != null) {
            return Formatter;
        }
        if (formattersDirectory) {
            Formatter = loadFormatter(formattersDirectory, camelizedName);
            if (Formatter) {
                return Formatter;
            }
        }
        return loadFormatterModule(name);
    }
    else {
        throw new Error("Name of type " + typeof name + " is not supported.");
    }
}
exports.findFormatter = findFormatter;
function loadFormatter() {
    var paths = [];
    for (var _i = 0; _i < arguments.length; _i++) {
        paths[_i - 0] = arguments[_i];
    }
    var formatterPath = paths.reduce(function (p, c) { return path.join(p, c); }, "");
    var fullPath = path.resolve(moduleDirectory, formatterPath);
    if (fs.existsSync(fullPath + ".js")) {
        var formatterModule = require(fullPath);
        return formatterModule.Formatter;
    }
    return undefined;
}
function loadFormatterModule(name) {
    var src;
    try {
        src = require.resolve(name);
    }
    catch (e) {
        return undefined;
    }
    return require(src).Formatter;
}
