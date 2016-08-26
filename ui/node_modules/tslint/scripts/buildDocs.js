"use strict";
var fs = require("fs");
var glob = require("glob");
var yaml = require("js-yaml");
var path = require("path");
var DOCS_DIR = "../docs";
var DOCS_RULE_DIR = path.join(DOCS_DIR, "rules");
var rulePaths = glob.sync("../lib/rules/*Rule.js");
var rulesJson = [];
for (var _i = 0, rulePaths_1 = rulePaths; _i < rulePaths_1.length; _i++) {
    var rulePath = rulePaths_1[_i];
    var ruleModule = require(rulePath);
    var Rule = ruleModule.Rule;
    if (Rule != null && Rule.metadata != null) {
        var metadata = Rule.metadata;
        var fileData_1 = generateRuleFile(metadata);
        var fileDirectory = path.join(DOCS_RULE_DIR, metadata.ruleName);
        if (!fs.existsSync(fileDirectory)) {
            fs.mkdirSync(fileDirectory);
        }
        fs.writeFileSync(path.join(fileDirectory, "index.html"), fileData_1);
        rulesJson.push(metadata);
    }
}
var fileData = JSON.stringify(rulesJson, undefined, 2);
fs.writeFileSync(path.join(DOCS_DIR, "_data", "rules.json"), fileData);
function generateRuleFile(metadata) {
    var yamlData = {};
    for (var _i = 0, _a = Object.keys(metadata); _i < _a.length; _i++) {
        var key = _a[_i];
        yamlData[key] = metadata[key];
    }
    yamlData.optionsJSON = JSON.stringify(metadata.options, undefined, 2);
    yamlData.layout = "rule";
    yamlData.title = "Rule: " + metadata.ruleName;
    return "---\n" + yaml.safeDump(yamlData, { lineWidth: 140 }) + "---";
}
