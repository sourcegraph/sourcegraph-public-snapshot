"use strict";
var ts = require("typescript");
var utils_1 = require("./utils");
function createLanguageServiceHost(fileName, source) {
    return {
        getCompilationSettings: function () { return utils_1.createCompilerOptions(); },
        getCurrentDirectory: function () { return ""; },
        getDefaultLibFileName: function () { return "lib.d.ts"; },
        getScriptFileNames: function () { return [fileName]; },
        getScriptSnapshot: function (name) { return ts.ScriptSnapshot.fromString(name === fileName ? source : ""); },
        getScriptVersion: function () { return "1"; },
        log: function () { },
    };
}
exports.createLanguageServiceHost = createLanguageServiceHost;
function createLanguageService(fileName, source) {
    var languageServiceHost = createLanguageServiceHost(fileName, source);
    return ts.createLanguageService(languageServiceHost);
}
exports.createLanguageService = createLanguageService;
