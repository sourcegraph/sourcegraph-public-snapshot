"use strict";
var ts = require("typescript");
var configuration_1 = require("./configuration");
var enableDisableRules_1 = require("./enableDisableRules");
var formatterLoader_1 = require("./formatterLoader");
var typedRule_1 = require("./language/rule/typedRule");
var utils_1 = require("./language/utils");
var ruleLoader_1 = require("./ruleLoader");
var utils_2 = require("./utils");
var Linter = (function () {
    function Linter(fileName, source, options, program) {
        this.fileName = fileName;
        this.source = source;
        this.program = program;
        this.options = this.computeFullOptions(options);
    }
    Linter.createProgram = function (configFile, projectDirectory) {
        if (projectDirectory === undefined) {
            var lastSeparator = configFile.lastIndexOf("/");
            if (lastSeparator < 0) {
                projectDirectory = ".";
            }
            else {
                projectDirectory = configFile.substring(0, lastSeparator + 1);
            }
        }
        var config = ts.readConfigFile(configFile, ts.sys.readFile).config;
        var parsed = ts.parseJsonConfigFileContent(config, { readDirectory: ts.sys.readDirectory }, projectDirectory);
        var host = ts.createCompilerHost(parsed.options, true);
        var program = ts.createProgram(parsed.fileNames, parsed.options, host);
        return program;
    };
    Linter.getFileNames = function (program) {
        return program.getSourceFiles().map(function (s) { return s.fileName; }).filter(function (l) { return l.substr(-5) !== ".d.ts"; });
    };
    Linter.prototype.lint = function () {
        var failures = [];
        var sourceFile;
        if (this.program) {
            sourceFile = this.program.getSourceFile(this.fileName);
            if (!("resolvedModules" in sourceFile)) {
                throw new Error("Program must be type checked before linting");
            }
        }
        else {
            sourceFile = utils_1.getSourceFile(this.fileName, this.source);
        }
        if (sourceFile === undefined) {
            throw new Error("Invalid source file: " + this.fileName + ". Ensure that the files supplied to lint have a .ts or .tsx extension.");
        }
        var rulesWalker = new enableDisableRules_1.EnableDisableRulesWalker(sourceFile, {
            disabledIntervals: [],
            ruleName: "",
        });
        rulesWalker.walk(sourceFile);
        var enableDisableRuleMap = rulesWalker.enableDisableRuleMap;
        var rulesDirectories = this.options.rulesDirectory;
        var configuration = this.options.configuration.rules;
        var configuredRules = ruleLoader_1.loadRules(configuration, enableDisableRuleMap, rulesDirectories);
        var enabledRules = configuredRules.filter(function (r) { return r.isEnabled(); });
        for (var _i = 0, enabledRules_1 = enabledRules; _i < enabledRules_1.length; _i++) {
            var rule = enabledRules_1[_i];
            var ruleFailures = [];
            if (this.program && rule instanceof typedRule_1.TypedRule) {
                ruleFailures = rule.applyWithProgram(sourceFile, this.program);
            }
            else {
                ruleFailures = rule.apply(sourceFile);
            }
            for (var _a = 0, ruleFailures_1 = ruleFailures; _a < ruleFailures_1.length; _a++) {
                var ruleFailure = ruleFailures_1[_a];
                if (!this.containsRule(failures, ruleFailure)) {
                    failures.push(ruleFailure);
                }
            }
        }
        var formatter;
        var formattersDirectory = configuration_1.getRelativePath(this.options.formattersDirectory);
        var Formatter = formatterLoader_1.findFormatter(this.options.formatter, formattersDirectory);
        if (Formatter) {
            formatter = new Formatter();
        }
        else {
            throw new Error("formatter '" + this.options.formatter + "' not found");
        }
        var output = formatter.format(failures);
        return {
            failureCount: failures.length,
            failures: failures,
            format: this.options.formatter,
            output: output,
        };
    };
    Linter.prototype.containsRule = function (rules, rule) {
        return rules.some(function (r) { return r.equals(rule); });
    };
    Linter.prototype.computeFullOptions = function (options) {
        if (options === void 0) { options = {}; }
        if (typeof options !== "object") {
            throw new Error("Unknown Linter options type: " + typeof options);
        }
        var configuration = options.configuration, formatter = options.formatter, formattersDirectory = options.formattersDirectory, rulesDirectory = options.rulesDirectory;
        return {
            configuration: configuration || configuration_1.DEFAULT_CONFIG,
            formatter: formatter || "prose",
            formattersDirectory: formattersDirectory,
            rulesDirectory: utils_2.arrayify(rulesDirectory).concat(utils_2.arrayify(configuration.rulesDirectory)),
        };
    };
    Linter.VERSION = "3.15.1";
    Linter.findConfiguration = configuration_1.findConfiguration;
    Linter.findConfigurationPath = configuration_1.findConfigurationPath;
    Linter.getRulesDirectories = configuration_1.getRulesDirectories;
    Linter.loadConfigurationFromPath = configuration_1.loadConfigurationFromPath;
    return Linter;
}());
module.exports = Linter;
