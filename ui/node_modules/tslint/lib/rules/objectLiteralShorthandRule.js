"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var ts = require("typescript");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var objectLiteralShorthandWalker = new ObjectLiteralShorthandWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(objectLiteralShorthandWalker);
    };
    Rule.metadata = {
        ruleName: "object-literal-shorthand",
        description: "Enforces use of ES6 object literal shorthand when possible.",
        options: null,
        optionExamples: ["true"],
        type: "style",
    };
    Rule.LONGHAND_PROPERTY = "Expected property shorthand in object literal.";
    Rule.LONGHAND_METHOD = "Expected method shorthand in object literal.";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var ObjectLiteralShorthandWalker = (function (_super) {
    __extends(ObjectLiteralShorthandWalker, _super);
    function ObjectLiteralShorthandWalker() {
        _super.apply(this, arguments);
    }
    ObjectLiteralShorthandWalker.prototype.visitPropertyAssignment = function (node) {
        var name = node.name;
        var value = node.initializer;
        if (name.kind === ts.SyntaxKind.Identifier &&
            value.kind === ts.SyntaxKind.Identifier &&
            name.getText() === value.getText()) {
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.LONGHAND_PROPERTY));
        }
        if (value.kind === ts.SyntaxKind.FunctionExpression) {
            var fnNode = value;
            if (fnNode.name) {
                return;
            }
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.LONGHAND_METHOD));
        }
        _super.prototype.visitPropertyAssignment.call(this, node);
    };
    return ObjectLiteralShorthandWalker;
}(Lint.RuleWalker));
