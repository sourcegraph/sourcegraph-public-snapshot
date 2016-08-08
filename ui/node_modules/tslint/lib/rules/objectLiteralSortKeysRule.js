"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new ObjectLiteralSortKeysWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "object-literal-sort-keys",
        description: "Requires keys in object literals to be sorted alphabetically",
        rationale: "Useful in preventing merge conflicts",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "maintainability",
    };
    Rule.FAILURE_STRING_FACTORY = function (name) { return ("The key '" + name + "' is not sorted alphabetically"); };
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var ObjectLiteralSortKeysWalker = (function (_super) {
    __extends(ObjectLiteralSortKeysWalker, _super);
    function ObjectLiteralSortKeysWalker() {
        _super.apply(this, arguments);
        this.lastSortedKeyStack = [];
        this.sortedStateStack = [];
    }
    ObjectLiteralSortKeysWalker.prototype.visitObjectLiteralExpression = function (node) {
        this.lastSortedKeyStack.push("");
        this.sortedStateStack.push(true);
        _super.prototype.visitObjectLiteralExpression.call(this, node);
        this.lastSortedKeyStack.pop();
        this.sortedStateStack.pop();
    };
    ObjectLiteralSortKeysWalker.prototype.visitPropertyAssignment = function (node) {
        var sortedState = this.sortedStateStack[this.sortedStateStack.length - 1];
        if (sortedState) {
            var lastSortedKey = this.lastSortedKeyStack[this.lastSortedKeyStack.length - 1];
            var keyNode = node.name;
            if (keyNode.kind === ts.SyntaxKind.Identifier) {
                var key = keyNode.text;
                if (key < lastSortedKey) {
                    var failureString = Rule.FAILURE_STRING_FACTORY(key);
                    this.addFailure(this.createFailure(keyNode.getStart(), keyNode.getWidth(), failureString));
                    this.sortedStateStack[this.sortedStateStack.length - 1] = false;
                }
                else {
                    this.lastSortedKeyStack[this.lastSortedKeyStack.length - 1] = key;
                }
            }
        }
        _super.prototype.visitPropertyAssignment.call(this, node);
    };
    return ObjectLiteralSortKeysWalker;
}(Lint.RuleWalker));
