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
        return this.applyWithWalker(new NoDuplicateKeyWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-duplicate-key",
        description: "Disallows duplicate keys in object literals.",
        rationale: (_a = ["\n            There is no good reason to define an object literal with the same key twice.\n            This rule is now implemented in the TypeScript compiler and does not need to be used."], _a.raw = ["\n            There is no good reason to define an object literal with the same key twice.\n            This rule is now implemented in the TypeScript compiler and does not need to be used."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING_FACTORY = function (name) { return ("Duplicate key '" + name + "'"); };
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoDuplicateKeyWalker = (function (_super) {
    __extends(NoDuplicateKeyWalker, _super);
    function NoDuplicateKeyWalker() {
        _super.apply(this, arguments);
        this.objectKeysStack = [];
    }
    NoDuplicateKeyWalker.prototype.visitObjectLiteralExpression = function (node) {
        this.objectKeysStack.push(Object.create(null));
        _super.prototype.visitObjectLiteralExpression.call(this, node);
        this.objectKeysStack.pop();
    };
    NoDuplicateKeyWalker.prototype.visitPropertyAssignment = function (node) {
        var objectKeys = this.objectKeysStack[this.objectKeysStack.length - 1];
        var keyNode = node.name;
        if (keyNode.kind === ts.SyntaxKind.Identifier) {
            var key = keyNode.text;
            if (objectKeys[key]) {
                var failureString = Rule.FAILURE_STRING_FACTORY(key);
                this.addFailure(this.createFailure(keyNode.getStart(), keyNode.getWidth(), failureString));
            }
            else {
                objectKeys[key] = true;
            }
        }
        _super.prototype.visitPropertyAssignment.call(this, node);
    };
    return NoDuplicateKeyWalker;
}(Lint.RuleWalker));
