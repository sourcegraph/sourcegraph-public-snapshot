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
        return this.applyWithWalker(new AdjacentOverloadSignaturesWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "adjacent-overload-signatures",
        description: "Enforces function overloads to be consecutive.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "typescript",
    };
    Rule.FAILURE_STRING_FACTORY = function (name) { return ("All '" + name + "' signatures should be adjacent"); };
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var AdjacentOverloadSignaturesWalker = (function (_super) {
    __extends(AdjacentOverloadSignaturesWalker, _super);
    function AdjacentOverloadSignaturesWalker() {
        _super.apply(this, arguments);
    }
    AdjacentOverloadSignaturesWalker.prototype.visitInterfaceDeclaration = function (node) {
        this.checkNode(node);
        _super.prototype.visitInterfaceDeclaration.call(this, node);
    };
    AdjacentOverloadSignaturesWalker.prototype.visitTypeLiteral = function (node) {
        this.checkNode(node);
        _super.prototype.visitTypeLiteral.call(this, node);
    };
    AdjacentOverloadSignaturesWalker.prototype.checkNode = function (node) {
        var last = undefined;
        var seen = {};
        for (var _i = 0, _a = node.members; _i < _a.length; _i++) {
            var member = _a[_i];
            if (member.name !== undefined) {
                var methodName = getTextOfPropertyName(member.name);
                if (methodName !== undefined) {
                    if (seen[methodName] && last !== methodName) {
                        this.addFailure(this.createFailure(member.getStart(), member.getWidth(), Rule.FAILURE_STRING_FACTORY(methodName)));
                    }
                    last = methodName;
                    seen[methodName] = true;
                }
            }
            else {
                last = undefined;
            }
        }
    };
    return AdjacentOverloadSignaturesWalker;
}(Lint.RuleWalker));
function isLiteralExpression(node) {
    return node.kind === ts.SyntaxKind.StringLiteral || node.kind === ts.SyntaxKind.NumericLiteral;
}
function getTextOfPropertyName(name) {
    switch (name.kind) {
        case ts.SyntaxKind.Identifier:
            return name.text;
        case ts.SyntaxKind.ComputedPropertyName:
            var expression = name.expression;
            if (isLiteralExpression(expression)) {
                return expression.text;
            }
        default:
            if (isLiteralExpression(name)) {
                return name.text;
            }
    }
}
