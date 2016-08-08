"use strict";
var fs = require("fs");
var path = require("path");
var underscore_string_1 = require("underscore.string");
var configuration_1 = require("./configuration");
var moduleDirectory = path.dirname(module.filename);
var CORE_RULES_DIRECTORY = path.resolve(moduleDirectory, ".", "rules");
function loadRules(ruleConfiguration, enableDisableRuleMap, rulesDirectories) {
    var rules = [];
    var notFoundRules = [];
    for (var ruleName in ruleConfiguration) {
        if (ruleConfiguration.hasOwnProperty(ruleName)) {
            var ruleValue = ruleConfiguration[ruleName];
            var Rule = findRule(ruleName, rulesDirectories);
            if (Rule == null) {
                notFoundRules.push(ruleName);
            }
            else {
                var all = "all";
                var allList = (all in enableDisableRuleMap ? enableDisableRuleMap[all] : []);
                var ruleSpecificList = (ruleName in enableDisableRuleMap ? enableDisableRuleMap[ruleName] : []);
                var disabledIntervals = buildDisabledIntervalsFromSwitches(ruleSpecificList, allList);
                rules.push(new Rule(ruleName, ruleValue, disabledIntervals));
            }
        }
    }
    if (notFoundRules.length > 0) {
        var ERROR_MESSAGE = "\n            Could not find implementations for the following rules specified in the configuration:\n            " + notFoundRules.join("\n") + "\n            Try upgrading TSLint and/or ensuring that you have all necessary custom rules installed.\n            If TSLint was recently upgraded, you may have old rules configured which need to be cleaned up.\n        ";
        throw new Error(ERROR_MESSAGE);
    }
    else {
        return rules;
    }
}
exports.loadRules = loadRules;
function findRule(name, rulesDirectories) {
    var camelizedName = transformName(name);
    var Rule = loadRule(CORE_RULES_DIRECTORY, camelizedName);
    if (Rule != null) {
        return Rule;
    }
    var directories = configuration_1.getRulesDirectories(rulesDirectories);
    for (var _i = 0, directories_1 = directories; _i < directories_1.length; _i++) {
        var rulesDirectory = directories_1[_i];
        if (rulesDirectory != null) {
            Rule = loadRule(rulesDirectory, camelizedName);
            if (Rule != null) {
                return Rule;
            }
        }
    }
    return undefined;
}
exports.findRule = findRule;
function transformName(name) {
    var nameMatch = name.match(/^([-_]*)(.*?)([-_]*)$/);
    if (nameMatch == null) {
        return name + "Rule";
    }
    return nameMatch[1] + underscore_string_1.camelize(nameMatch[2]) + nameMatch[3] + "Rule";
}
function loadRule(directory, ruleName) {
    var fullPath = path.join(directory, ruleName);
    if (fs.existsSync(fullPath + ".js")) {
        var ruleModule = require(fullPath);
        if (ruleModule && ruleModule.Rule) {
            return ruleModule.Rule;
        }
    }
    return undefined;
}
function buildDisabledIntervalsFromSwitches(ruleSpecificList, allList) {
    var isCurrentlyDisabled = false;
    var disabledStartPosition;
    var disabledIntervalList = [];
    var i = 0;
    var j = 0;
    while (i < ruleSpecificList.length || j < allList.length) {
        var ruleSpecificTopPositon = (i < ruleSpecificList.length ? ruleSpecificList[i].position : Infinity);
        var allTopPositon = (j < allList.length ? allList[j].position : Infinity);
        var newPositionToCheck = void 0;
        if (ruleSpecificTopPositon < allTopPositon) {
            newPositionToCheck = ruleSpecificList[i];
            i++;
        }
        else {
            newPositionToCheck = allList[j];
            j++;
        }
        if (newPositionToCheck.isEnabled === isCurrentlyDisabled) {
            if (!isCurrentlyDisabled) {
                disabledStartPosition = newPositionToCheck.position;
                isCurrentlyDisabled = true;
            }
            else {
                disabledIntervalList.push({
                    endPosition: newPositionToCheck.position,
                    startPosition: disabledStartPosition,
                });
                isCurrentlyDisabled = false;
            }
        }
    }
    if (isCurrentlyDisabled) {
        disabledIntervalList.push({
            endPosition: Infinity,
            startPosition: disabledStartPosition,
        });
    }
    return disabledIntervalList;
}
