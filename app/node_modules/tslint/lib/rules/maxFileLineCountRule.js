"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.isEnabled = function () {
        if (_super.prototype.isEnabled.call(this)) {
            var option = this.getOptions().ruleArguments[0];
            if (typeof option === "number" && option > 0) {
                return true;
            }
        }
        return false;
    };
    Rule.prototype.apply = function (sourceFile) {
        var ruleFailures = [];
        var lineLimit = this.getOptions().ruleArguments[0];
        var lineCount = sourceFile.getLineStarts().length;
        var disabledIntervals = this.getOptions().disabledIntervals;
        if (lineCount > lineLimit && disabledIntervals.length === 0) {
            var errorString = Rule.FAILURE_STRING_FACTORY(lineCount, lineLimit);
            ruleFailures.push(new Lint.RuleFailure(sourceFile, 0, 1, errorString, this.getOptions().ruleName));
        }
        return ruleFailures;
    };
    Rule.metadata = {
        ruleName: "max-file-line-count",
        description: "Requires files to remain under a certain number of lines",
        rationale: (_a = ["\n            Limiting the number of lines allowed in a file allows files to remain small, \n            single purpose, and maintainable."], _a.raw = ["\n            Limiting the number of lines allowed in a file allows files to remain small, \n            single purpose, and maintainable."], Lint.Utils.dedent(_a)),
        optionsDescription: "An integer indicating the maximum number of lines.",
        options: {
            type: "number",
            minimum: "1",
        },
        optionExamples: ["[true, 300]"],
        type: "maintainability",
    };
    Rule.FAILURE_STRING_FACTORY = function (lineCount, lineLimit) {
        var msg = "This file has " + lineCount + " lines, which exceeds the maximum of " + lineLimit + " lines allowed. ";
        msg += "Consider breaking this file up into smaller parts";
        return msg;
    };
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
