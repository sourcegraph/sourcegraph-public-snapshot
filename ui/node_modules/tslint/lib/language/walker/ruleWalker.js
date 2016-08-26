"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var rule_1 = require("../rule/rule");
var utils_1 = require("../utils");
var syntaxWalker_1 = require("./syntaxWalker");
var RuleWalker = (function (_super) {
    __extends(RuleWalker, _super);
    function RuleWalker(sourceFile, options) {
        _super.call(this);
        this.sourceFile = sourceFile;
        this.position = 0;
        this.failures = [];
        this.options = options.ruleArguments;
        this.limit = this.sourceFile.getFullWidth();
        this.disabledIntervals = options.disabledIntervals;
        this.ruleName = options.ruleName;
    }
    RuleWalker.prototype.getSourceFile = function () {
        return this.sourceFile;
    };
    RuleWalker.prototype.getFailures = function () {
        return this.failures;
    };
    RuleWalker.prototype.getLimit = function () {
        return this.limit;
    };
    RuleWalker.prototype.getOptions = function () {
        return this.options;
    };
    RuleWalker.prototype.hasOption = function (option) {
        if (this.options) {
            return this.options.indexOf(option) !== -1;
        }
        else {
            return false;
        }
    };
    RuleWalker.prototype.skip = function (node) {
        this.position += node.getFullWidth();
    };
    RuleWalker.prototype.createFailure = function (start, width, failure, fix) {
        var from = (start > this.limit) ? this.limit : start;
        var to = ((start + width) > this.limit) ? this.limit : (start + width);
        return new rule_1.RuleFailure(this.sourceFile, from, to, failure, this.ruleName, fix);
    };
    RuleWalker.prototype.addFailure = function (failure) {
        if (!this.existsFailure(failure) && !utils_1.doesIntersect(failure, this.disabledIntervals)) {
            this.failures.push(failure);
        }
    };
    RuleWalker.prototype.createReplacement = function (start, length, text) {
        return new rule_1.Replacement(start, length, text);
    };
    RuleWalker.prototype.appendText = function (start, text) {
        return this.createReplacement(start, 0, text);
    };
    RuleWalker.prototype.deleteText = function (start, length) {
        return this.createReplacement(start, length, "");
    };
    RuleWalker.prototype.existsFailure = function (failure) {
        return this.failures.some(function (f) { return f.equals(failure); });
    };
    return RuleWalker;
}(syntaxWalker_1.SyntaxWalker));
exports.RuleWalker = RuleWalker;
