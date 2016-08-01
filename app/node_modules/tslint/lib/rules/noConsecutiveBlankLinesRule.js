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
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new NoConsecutiveBlankLinesWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-consecutive-blank-lines",
        description: "Disallows more than one blank line in a row.",
        rationale: "Helps maintain a readable style in your codebase.",
        optionsDescription: "Not configurable.",
        options: {},
        optionExamples: ["true"],
        type: "style",
    };
    Rule.FAILURE_STRING = "Consecutive blank lines are forbidden";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoConsecutiveBlankLinesWalker = (function (_super) {
    __extends(NoConsecutiveBlankLinesWalker, _super);
    function NoConsecutiveBlankLinesWalker() {
        _super.apply(this, arguments);
    }
    NoConsecutiveBlankLinesWalker.prototype.visitSourceFile = function (node) {
        var _this = this;
        _super.prototype.visitSourceFile.call(this, node);
        var sourceFileText = node.getFullText();
        var soureFileLines = sourceFileText.split(/\n/);
        var blankLineIndexes = [];
        soureFileLines.forEach(function (txt, i) {
            if (txt.trim() === "") {
                blankLineIndexes.push(i);
            }
        });
        var sequences = [];
        var lastVal = -2;
        for (var _i = 0, blankLineIndexes_1 = blankLineIndexes; _i < blankLineIndexes_1.length; _i++) {
            var line = blankLineIndexes_1[_i];
            line > lastVal + 1 ? sequences.push([line]) : sequences[sequences.length - 1].push(line);
            lastVal = line;
        }
        sequences
            .filter(function (arr) { return arr.length > 1; }).map(function (arr) { return arr[0]; })
            .forEach(function (startLineNum) {
            var startCharPos = node.getPositionOfLineAndCharacter(startLineNum + 1, 0);
            _this.addFailure(_this.createFailure(startCharPos, 1, Rule.FAILURE_STRING));
        });
    };
    return NoConsecutiveBlankLinesWalker;
}(Lint.SkippableTokenAwareRuleWalker));
